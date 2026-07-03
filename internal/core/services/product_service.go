package services

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

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

// BulkUpdatePrice aplica un porcentaje al base_price de un conjunto de productos.
// El conjunto se resuelve así: si ids no está vacío, esos productos; si no y
// category != "", toda esa categoría; si no, todos los productos. Devuelve cuántos
// productos actualizó. No toca los price_adjustment de las variantes.
func (s *ProductService) BulkUpdatePrice(ctx context.Context, percent float64, ids []primitive.ObjectID, category string) (int, error) {
	if percent == 0 {
		return 0, errors.New("el porcentaje no puede ser 0")
	}
	if percent < -90 || percent > 1000 {
		return 0, errors.New("porcentaje fuera de rango (-90 a 1000)")
	}

	// Resolver el conjunto de productos.
	var products []*domain.Product
	switch {
	case len(ids) > 0:
		for _, id := range ids {
			p, err := s.productRepository.GetByID(ctx, id)
			if err != nil || p == nil {
				continue // ignorar IDs inválidos / inexistentes
			}
			products = append(products, p)
		}
	case category != "":
		list, err := s.productRepository.GetByCategory(ctx, category, 0, 100000)
		if err != nil {
			return 0, err
		}
		products = list
	default:
		list, err := s.productRepository.List(ctx, map[string]interface{}{}, 0, 100000)
		if err != nil {
			return 0, err
		}
		products = list
	}

	factor := 1 + percent/100
	updated := 0
	for _, p := range products {
		newPrice := math.Round(p.BasePrice * factor)
		if newPrice < 0 {
			newPrice = 0
		}
		p.BasePrice = newPrice
		p.UpdatedAt = time.Now()
		if err := s.productRepository.Update(ctx, p); err != nil {
			log.Printf("BulkUpdatePrice: no se pudo actualizar producto %s: %v", p.ID.Hex(), err)
			continue
		}
		updated++
	}
	return updated, nil
}
