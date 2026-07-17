package ports

import (
	"context"
	"time"

	"ecommerce-pooled/internal/core/domain"
)

// EventRepository define el contrato para persistencia de eventos de analytics.
type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	CountByType(ctx context.Context, eventType string, from, to time.Time) (int64, error)
	GroupByDay(ctx context.Context, eventType string, from, to time.Time) ([]domain.DayCount, error)
	TopItems(ctx context.Context, eventType, idField, nameField string, limit int, from, to time.Time) ([]domain.ItemStat, error)
	CountUniqueSessions(ctx context.Context, eventType string, from, to time.Time) (int64, error)
	// CountUniqueSessionsAny cuenta session_id distintos en el rango, sin filtrar por tipo.
	CountUniqueSessionsAny(ctx context.Context, from, to time.Time) (int64, error)
	// SessionsByLocation agrupa sesiones únicas por país/ciudad en el rango dado.
	SessionsByLocation(ctx context.Context, from, to time.Time) ([]domain.LocationStat, error)
}
