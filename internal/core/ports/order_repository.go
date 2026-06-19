package ports

import (
	"context"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderRepository define el contrato para las operaciones de persistencia de órdenes
type OrderRepository interface {
	// Create crea una nueva orden en la base de datos
	Create(ctx context.Context, order *domain.Order) error

	// GetByID obtiene una orden por su ID
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)

	// GetByUserID obtiene las órdenes de un usuario específico
	GetByUserID(ctx context.Context, userID primitive.ObjectID, skip int64, limit int64) ([]*domain.Order, error)

	// Update actualiza una orden existente
	Update(ctx context.Context, order *domain.Order) error

	// UpdateStatus actualiza el estado de una orden
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error

	// Delete elimina una orden por su ID
	Delete(ctx context.Context, id primitive.ObjectID) error

	// List obtiene una lista de todas las órdenes con paginación
	List(ctx context.Context, skip int64, limit int64) ([]*domain.Order, error)

	// GetByStatus obtiene órdenes filtradas por estado
	GetByStatus(ctx context.Context, status string, skip int64, limit int64) ([]*domain.Order, error)

	// CountByStatusInPeriod cuenta órdenes con alguno de los estados dados, creadas entre from y to.
	CountByStatusInPeriod(ctx context.Context, statuses []string, from, to time.Time) (int64, error)

	// FindPendingOlderThan devuelve órdenes en estado "pending" creadas antes de `before`.
	FindPendingOlderThan(ctx context.Context, before time.Time) ([]*domain.Order, error)

	// Count devuelve el total de órdenes, opcionalmente filtrado por status.
	Count(ctx context.Context, status string) (int64, error)
}
