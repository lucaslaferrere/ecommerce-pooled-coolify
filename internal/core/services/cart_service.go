package services

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
)

// CartService valida ítems del carrito contra la base de datos
type CartService struct {
	productRepository ports.ProductRepository
}

// NewCartService crea una nueva instancia de CartService
func NewCartService(productRepository ports.ProductRepository) *CartService {
	return &CartService{productRepository: productRepository}
}

// ValidatedCart contiene los ítems validados y enriquecidos con precios
type ValidatedCart struct {
	Items []domain.CartItem
	Total float64
}

// ValidateCart verifica que todos los productos/variantes existan y tengan stock suficiente.
// Retorna los ítems con UnitPrice calculado (basePrice + priceAdjustment) y el total.
func (s *CartService) ValidateCart(ctx context.Context, items []domain.CartItem) (*ValidatedCart, error) {
	if len(items) == 0 {
		return nil, errors.New("el carrito no puede estar vacío")
	}

	validated := make([]domain.CartItem, len(items))
	var total float64

	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, errors.New("la cantidad debe ser mayor a 0")
		}

		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, ErrProductNotFound
		}

		var unitPrice float64
		if item.VariantSKU == "" || len(product.Variants) == 0 {
			// Producto sin variante: usa base_price directamente
			unitPrice = product.BasePrice
		} else {
			var found *domain.Variant
			for j := range product.Variants {
				if product.Variants[j].SKU == item.VariantSKU {
					found = &product.Variants[j]
					break
				}
			}
			if found == nil {
				return nil, errors.New("variante no encontrada: " + item.VariantSKU)
			}
			if found.Stock < item.Quantity {
				return nil, ErrInsufficientStock
			}
			unitPrice = product.BasePrice + found.PriceAdjustment
		}

		validated[i] = domain.CartItem{
			ProductID:  item.ProductID,
			VariantSKU: item.VariantSKU,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
		}
		total += unitPrice * float64(item.Quantity)
	}

	return &ValidatedCart{Items: validated, Total: total}, nil
}
