package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductRepository define el contrato para las operaciones de persistencia de productos
type ProductRepository interface {
	// Create crea un nuevo producto en la base de datos
	Create(ctx context.Context, product *domain.Product) error

	// GetByID obtiene un producto por su ID
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error)

	// Update actualiza un producto existente
	Update(ctx context.Context, product *domain.Product) error

	// Delete elimina un producto por su ID
	Delete(ctx context.Context, id primitive.ObjectID) error

	// List obtiene una lista de productos con filtros opcionales y paginación
	List(ctx context.Context, filter map[string]interface{}, skip int64, limit int64) ([]*domain.Product, error)

	// GetByCategory obtiene productos por categoría
	GetByCategory(ctx context.Context, category string, skip int64, limit int64) ([]*domain.Product, error)

	// GetByBrand obtiene productos por marca
	GetByBrand(ctx context.Context, brand string, skip int64, limit int64) ([]*domain.Product, error)

	// UpdateVariantStock actualiza el stock de una variante de un producto
	UpdateVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, newStock int) error

	// DecrementVariantStock reduce el stock de forma atómica; falla si el stock es insuficiente
	DecrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error

	// IncrementVariantStock incrementa el stock de una variante (usado para rollback)
	IncrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error
}
