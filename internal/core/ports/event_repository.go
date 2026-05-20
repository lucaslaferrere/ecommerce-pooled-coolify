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
}
