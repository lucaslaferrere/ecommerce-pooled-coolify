package services

import (
	"context"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
)

type EventService struct {
	repo      ports.EventRepository
	orderRepo ports.OrderRepository
}

func NewEventService(repo ports.EventRepository, orderRepo ports.OrderRepository) *EventService {
	return &EventService{repo: repo, orderRepo: orderRepo}
}

func (s *EventService) Track(ctx context.Context, event *domain.Event) error {
	if event.Type == "" {
		return nil
	}
	event.CreatedAt = time.Now()
	return s.repo.Create(ctx, event)
}

func (s *EventService) GetAnalytics(ctx context.Context, period string) (*domain.AnalyticsReport, error) {
	from, to := periodBounds(period)

	cartAdds, _ := s.repo.CountByType(ctx, "cart_add", from, to)
	checkoutStarts, _ := s.repo.CountByType(ctx, "checkout_start", from, to)
	productViews, _ := s.repo.CountByType(ctx, "product_view", from, to)

	paidStatuses := []string{"paid", "processing", "shipped", "delivered"}
	paidOrders, _ := s.orderRepo.CountByStatusInPeriod(ctx, paidStatuses, from, to)

	byDay, _ := s.repo.GroupByDay(ctx, "cart_add", from, to)

	viewItems, _ := s.repo.TopItems(ctx, "product_view", "product_id", "product_name", 10, from, to)
	cartItems, _ := s.repo.TopItems(ctx, "cart_add", "item_id", "item_name", 10, from, to)

	var convRate float64
	if checkoutStarts > 0 {
		convRate = float64(paidOrders) / float64(checkoutStarts)
		if convRate > 1 {
			convRate = 1
		}
	}

	cartMap := make(map[string]int64, len(cartItems))
	for _, ci := range cartItems {
		cartMap[ci.ID] = ci.Count
	}

	topN := len(viewItems)
	if topN > 5 {
		topN = 5
	}
	topProducts := make([]domain.ProductStat, 0, topN)
	for _, vi := range viewItems[:topN] {
		topProducts = append(topProducts, domain.ProductStat{
			ID:       vi.ID,
			Name:     vi.Name,
			Views:    vi.Count,
			CartAdds: cartMap[vi.ID],
		})
	}

	return &domain.AnalyticsReport{
		Period:         period,
		CartAdds:       cartAdds,
		CheckoutStarts: checkoutStarts,
		ProductViews:   productViews,
		PaidOrders:     paidOrders,
		ConversionRate: convRate,
		ByDay:          byDay,
		TopProducts:    topProducts,
	}, nil
}

func periodBounds(period string) (from, to time.Time) {
	to = time.Now()
	switch period {
	case "30d":
		from = to.AddDate(0, 0, -30)
	default:
		from = to.AddDate(0, 0, -7)
	}
	return
}
