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
	kitRepository     ports.KitRepository
}

// NewCartService crea una nueva instancia de CartService
func NewCartService(productRepository ports.ProductRepository, kitRepository ports.KitRepository) *CartService {
	return &CartService{productRepository: productRepository, kitRepository: kitRepository}
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

		var unitPrice float64
		var itemName string
		sku := item.VariantSKU

		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		if product != nil {
			// Es un producto normal
			if len(product.Variants) == 0 {
				if product.Stock < item.Quantity {
					return nil, ErrInsufficientStock
				}
				unitPrice = product.BasePrice
			} else if sku == "" {
				first := &product.Variants[0]
				for j := range product.Variants {
					if product.Variants[j].Stock >= item.Quantity {
						first = &product.Variants[j]
						break
					}
				}
				if first.Stock < item.Quantity {
					return nil, ErrInsufficientStock
				}
				sku = first.SKU
				unitPrice = product.BasePrice + first.PriceAdjustment
			} else {
				var found *domain.Variant
				for j := range product.Variants {
					if product.Variants[j].SKU == sku {
						found = &product.Variants[j]
						break
					}
				}
				if found == nil {
					return nil, errors.New("variante no encontrada: " + sku)
				}
				if found.Stock < item.Quantity {
					return nil, ErrInsufficientStock
				}
				unitPrice = product.BasePrice + found.PriceAdjustment
			}
			itemName = product.Name
		} else {
			// Fallback: buscar como kit
			kit, err := s.kitRepository.GetByID(ctx, item.ProductID)
			if err != nil || kit == nil {
				return nil, ErrProductNotFound
			}
			unitPrice = kit.Price
			itemName = kit.Name
		}

		validated[i] = domain.CartItem{
			ProductID:  item.ProductID,
			Name:       itemName,
			VariantSKU: sku,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
		}
		total += unitPrice * float64(item.Quantity)
	}

	return &ValidatedCart{Items: validated, Total: total}, nil
}
