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

	// CountNewVsReturning agrupa órdenes (solo estados pagados) por user_id:
	// devuelve (nuevos, total), donde nuevos = clientes distintos (1ra orden c/u)
	// y total = todas las órdenes pagadas.
	CountNewVsReturning(ctx context.Context) (nuevos, total int64, err error)

	// SalesByPeriod agrega ventas pagadas por mes ("month") o día ("day") en el rango dado.
	SalesByPeriod(ctx context.Context, from, to time.Time, granularity string) ([]domain.SalesBucket, error)

	// SalesTotals devuelve totales del período: facturación y órdenes pagadas,
	// más el conteo de todas las órdenes (cualquier estado).
	SalesTotals(ctx context.Context, from, to time.Time) (domain.SalesTotals, error)

	// SalesByCategory agrega facturación pagada por categoría de producto.
	SalesByCategory(ctx context.Context, from, to time.Time) ([]domain.CategoryRevenue, error)

	// TopSellingProducts devuelve los productos más vendidos por unidades.
	TopSellingProducts(ctx context.Context, from, to time.Time, limit int) ([]domain.ProductSales, error)
}
