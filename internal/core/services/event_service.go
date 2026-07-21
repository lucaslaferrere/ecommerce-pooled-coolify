package services

import (
	"context"
	"log"
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

// GetSalesSeries devuelve la evolución de ventas pagadas agregada por mes o día.
func (s *EventService) GetSalesSeries(ctx context.Context, from, to time.Time, granularity string) ([]domain.SalesBucket, error) {
	if granularity != "day" {
		granularity = "month"
	}
	return s.orderRepo.SalesByPeriod(ctx, from, to, granularity)
}

// GetSalesOverview arma todas las métricas de venta del dashboard para un rango,
// más el rango previo (para los % de tendencia). Los errores parciales se loguean
// y se devuelve lo que se pudo calcular.
func (s *EventService) GetSalesOverview(ctx context.Context, from, to, prevFrom, prevTo time.Time, granularity string) (*domain.SalesOverview, error) {
	current, err := s.orderRepo.SalesTotals(ctx, from, to)
	if err != nil {
		log.Printf("overview totales actuales: %v", err)
	}
	previous, err := s.orderRepo.SalesTotals(ctx, prevFrom, prevTo)
	if err != nil {
		log.Printf("overview totales previos: %v", err)
	}
	series, err := s.GetSalesSeries(ctx, from, to, granularity)
	if err != nil {
		log.Printf("overview serie: %v", err)
	}
	byCategory, err := s.orderRepo.SalesByCategory(ctx, from, to)
	if err != nil {
		log.Printf("overview categorías: %v", err)
	}
	topProducts, err := s.orderRepo.TopSellingProducts(ctx, from, to, 5)
	if err != nil {
		log.Printf("overview top productos: %v", err)
	}

	return &domain.SalesOverview{
		Current:     current,
		Previous:    previous,
		Series:      series,
		ByCategory:  byCategory,
		TopProducts: topProducts,
	}, nil
}

// GetRealtimeSnapshot arma el estado inicial del dashboard en tiempo real.
// Los errores de cada agregación se loguean pero no abortan el snapshot: se
// devuelve lo que se pudo calcular (siempre 200 para el frontend).
func (s *EventService) GetRealtimeSnapshot(ctx context.Context, activeCount int) (*domain.RealtimeSnapshot, error) {
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		log.Printf("snapshot: no se pudo cargar timezone AR: %v — usando time.Local", err)
		loc = time.Local
	}
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	hourAgo := now.Add(-1 * time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	paidStatuses := []string{"paid", "processing", "shipped", "delivered"}

	carritos, err := s.repo.CountByType(ctx, "cart_add", todayStart, now)
	if err != nil {
		log.Printf("snapshot cart_add: %v", err)
	}
	enPago, err := s.repo.CountByType(ctx, "checkout_start", todayStart, now)
	if err != nil {
		log.Printf("snapshot checkout_start: %v", err)
	}
	comprasHoy, err := s.orderRepo.CountByStatusInPeriod(ctx, paidStatuses, todayStart, now)
	if err != nil {
		log.Printf("snapshot compras_hoy: %v", err)
	}
	ultimaHora, err := s.repo.CountUniqueSessionsAny(ctx, hourAgo, now)
	if err != nil {
		log.Printf("snapshot visitantes_ultima_hora: %v", err)
	}
	locations, err := s.repo.SessionsByLocation(ctx, dayAgo, now)
	if err != nil {
		log.Printf("snapshot sesiones_por_ubicacion: %v", err)
	}
	nuevos, total, err := s.orderRepo.CountNewVsReturning(ctx)
	if err != nil {
		log.Printf("snapshot nuevos_vs_recurrentes: %v", err)
	}

	return &domain.RealtimeSnapshot{
		VisitantesActivos:    activeCount,
		VisitantesUltimaHora: ultimaHora,
		CarritosActivos:      carritos,
		EnPago:               enPago,
		ComprasHoy:           comprasHoy,
		SesionesPorUbicacion: locations,
		NuevosVsRecurrentes:  domain.NewVsReturning{Nuevos: nuevos, Recurrentes: total - nuevos},
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
