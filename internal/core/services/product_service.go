package services

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductService contiene la lógica de negocio relacionada con productos
type ProductService struct {
	productRepository ports.ProductRepository
}

// NewProductService crea una nueva instancia de ProductService
func NewProductService(productRepository ports.ProductRepository) *ProductService {
	return &ProductService{
		productRepository: productRepository,
	}
}

// CreateProduct crea un nuevo producto
func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) error {
	// Validaciones de negocio
	if product.Name == "" {
		return errors.New("nombre de producto requerido")
	}

	if product.BasePrice <= 0 {
		return errors.New("precio base debe ser mayor a 0")
	}

	if len(product.Variants) == 0 {
		return errors.New("un producto debe tener al menos una variante")
	}

	return s.productRepository.Create(ctx, product)
}

// GetProduct obtiene un producto por ID
func (s *ProductService) GetProduct(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	product, err := s.productRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// UpdateProduct actualiza un producto
func (s *ProductService) UpdateProduct(ctx context.Context, product *domain.Product) error {
	if product.Name == "" {
		return errors.New("nombre de producto requerido")
	}

	if product.BasePrice <= 0 {
		return errors.New("precio base debe ser mayor a 0")
	}

	return s.productRepository.Update(ctx, product)
}

// DeleteProduct elimina un producto
func (s *ProductService) DeleteProduct(ctx context.Context, id primitive.ObjectID) error {
	return s.productRepository.Delete(ctx, id)
}

// GetProductsByCategory obtiene productos por categoría
func (s *ProductService) GetProductsByCategory(ctx context.Context, category string, skip int64, limit int64) ([]*domain.Product, error) {
	return s.productRepository.GetByCategory(ctx, category, skip, limit)
}

// ReduceVariantStock reduce el stock de una variante
func (s *ProductService) ReduceVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error {
	product, err := s.productRepository.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Encontrar la variante y verificar stock
	var variant *domain.Variant
	for i := range product.Variants {
		if product.Variants[i].SKU == sku {
			variant = &product.Variants[i]
			break
		}
	}

	if variant == nil {
		return errors.New("variante no encontrada")
	}

	if variant.Stock < quantity {
		return ErrInsufficientStock
	}

	newStock := variant.Stock - quantity
	return s.productRepository.UpdateVariantStock(ctx, productID, sku, newStock)
}

// ListProducts obtiene una lista de productos con paginación y filtros opcionales
func (s *ProductService) ListProducts(ctx context.Context, filter map[string]interface{}, skip int64, limit int64) ([]*domain.Product, error) {
	if skip < 0 {
		skip = 0
	}

	if limit <= 0 {
		limit = 10 // Límite por defecto
	}

	if limit > 100 {
		limit = 100 // Límite máximo
	}

	return s.productRepository.List(ctx, filter, skip, limit)
}

// GetProductsByBrand obtiene productos por marca
func (s *ProductService) GetProductsByBrand(ctx context.Context, brand string, skip int64, limit int64) ([]*domain.Product, error) {
	if brand == "" {
		return nil, errors.New("marca requerida")
	}

	return s.productRepository.GetByBrand(ctx, brand, skip, limit)
}

// UpdateVariantPrice actualiza el precio de una variante
func (s *ProductService) UpdateVariantPrice(ctx context.Context, productID primitive.ObjectID, sku string, newPriceAdjustment float64) error {
	product, err := s.productRepository.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	if product == nil {
		return ErrProductNotFound
	}

	// Encontrar y actualizar la variante
	found := false
	for i := range product.Variants {
		if product.Variants[i].SKU == sku {
			product.Variants[i].PriceAdjustment = newPriceAdjustment
			found = true
			break
		}
	}

	if !found {
		return errors.New("variante no encontrada")
	}

	return s.productRepository.Update(ctx, product)
}
