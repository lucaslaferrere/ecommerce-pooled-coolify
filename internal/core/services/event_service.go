package services

import (
	"context"
	"sort"
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

	// Merge views and cart_adds — items appear even if they only have one of the two
	type entry struct {
		Name     string
		Views    int64
		CartAdds int64
	}
	combined := make(map[string]*entry)
	for _, vi := range viewItems {
		e := &entry{Name: vi.Name, Views: vi.Count}
		combined[vi.ID] = e
	}
	for _, ci := range cartItems {
		if e, ok := combined[ci.ID]; ok {
			e.CartAdds = ci.Count
		} else {
			combined[ci.ID] = &entry{Name: ci.Name, CartAdds: ci.Count}
		}
	}

	type scored struct {
		id    string
		entry *entry
	}
	scoredList := make([]scored, 0, len(combined))
	for id, e := range combined {
		scoredList = append(scoredList, scored{id, e})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		si := scoredList[i].entry.Views + scoredList[i].entry.CartAdds
		sj := scoredList[j].entry.Views + scoredList[j].entry.CartAdds
		return si > sj
	})

	topN := 5
	if len(scoredList) < topN {
		topN = len(scoredList)
	}
	topProducts := make([]domain.ProductStat, 0, topN)
	for _, s := range scoredList[:topN] {
		topProducts = append(topProducts, domain.ProductStat{
			ID:       s.id,
			Name:     s.entry.Name,
			Views:    s.entry.Views,
			CartAdds: s.entry.CartAdds,
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

func (s *EventService) GetTrafficReport(ctx context.Context) (*domain.TrafficReport, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, 0, -30)

	viewsToday, _ := s.repo.CountByType(ctx, "page_view", todayStart, now)
	viewsWeek, _ := s.repo.CountByType(ctx, "page_view", weekAgo, now)
	viewsMonth, _ := s.repo.CountByType(ctx, "page_view", monthAgo, now)
	uniqueSessions, _ := s.repo.CountUniqueSessions(ctx, "page_view", monthAgo, now)

	rawPages, _ := s.repo.TopItems(ctx, "page_view", "url", "url", 10, monthAgo, now)
	pages := make([]domain.PageStat, 0, len(rawPages))
	for _, p := range rawPages {
		pages = append(pages, domain.PageStat{URL: p.ID, Count: p.Count})
	}

	return &domain.TrafficReport{
		ViewsToday:     viewsToday,
		ViewsWeek:      viewsWeek,
		ViewsMonth:     viewsMonth,
		UniqueSessions: uniqueSessions,
		TopPages:       pages,
	}, nil
}

func periodBounds(period string) (from, to time.Time) {
	to = time.Now()
	switch period {
	case "1d":
		// Hoy, desde las 00:00 hs.
		from = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	case "30d":
		from = to.AddDate(0, 0, -30)
	case "1y":
		// Año a la fecha, desde el 1 de enero.
		from = time.Date(to.Year(), 1, 1, 0, 0, 0, 0, to.Location())
	default:
		from = to.AddDate(0, 0, -7)
	}
	return
}
