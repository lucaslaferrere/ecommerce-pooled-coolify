package services

import (
	"context"
	"errors"
	"log"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderService contiene la lógica de negocio relacionada con órdenes
type OrderService struct {
	orderRepository   ports.OrderRepository
	productRepository ports.ProductRepository
}

// NewOrderService crea una nueva instancia de OrderService
func NewOrderService(orderRepository ports.OrderRepository, productRepository ports.ProductRepository) *OrderService {
	return &OrderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
	}
}

// CreateOrder crea una nueva orden
func (s *OrderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	// Validaciones de negocio
	if order.UserID.IsZero() {
		return errors.New("ID de usuario requerido")
	}

	if len(order.Items) == 0 {
		return errors.New("una orden debe tener al menos un item")
	}

	if order.Total <= 0 {
		return errors.New("total debe ser mayor a 0")
	}

	// Establecer estado inicial
	if order.Status == "" {
		order.Status = "pending"
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Verificar disponibilidad de productos y stock
	for _, item := range order.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			return err
		}

		if product == nil {
			return ErrProductNotFound
		}

		// Verificar disponibilidad de la variante
		found := false
		for _, variant := range product.Variants {
			if variant.SKU == item.VariantSKU {
				if variant.Stock < item.Quantity {
					return ErrInsufficientStock
				}
				found = true
				break
			}
		}

		if !found {
			return errors.New("variante no encontrada")
		}
	}

	return s.orderRepository.Create(ctx, order)
}

// GetOrder obtiene una orden por ID
func (s *OrderService) GetOrder(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// GetUserOrders obtiene todas las órdenes de un usuario
func (s *OrderService) GetUserOrders(ctx context.Context, userID primitive.ObjectID, skip int64, limit int64) ([]*domain.Order, error) {
	return s.orderRepository.GetByUserID(ctx, userID, skip, limit)
}

// UpdateOrderStatus actualiza el estado de una orden
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	// Validar estados válidos
	validStatuses := map[string]bool{
		"pending":    true,
		"paid":       true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
		"rejected":   true,
	}

	if !validStatuses[status] {
		return ErrInvalidOrderStatus
	}

	return s.orderRepository.UpdateStatus(ctx, id, status)
}

// CancelOrder cancela una orden y restaura el stock de sus items.
func (s *OrderService) CancelOrder(ctx context.Context, id primitive.ObjectID) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// Solo se pueden cancelar órdenes que están en estado pending o processing
	if order.Status != "pending" && order.Status != "processing" {
		return errors.New("solo se pueden cancelar órdenes en estado pending o processing")
	}

	for _, item := range order.Items {
		if item.ItemType == "kit" {
			continue
		}
		if item.VariantSKU == "" {
			_ = s.productRepository.IncrementProductStock(ctx, item.ProductID, item.Quantity)
		} else {
			_ = s.productRepository.IncrementVariantStock(ctx, item.ProductID, item.VariantSKU, item.Quantity)
		}
	}

	return s.orderRepository.UpdateStatus(ctx, id, "cancelled")
}

// DeleteOrder elimina una orden
func (s *OrderService) DeleteOrder(ctx context.Context, id primitive.ObjectID) error {
	return s.orderRepository.Delete(ctx, id)
}

// ListOrders obtiene todas las órdenes con paginación (solo admin)
func (s *OrderService) ListOrders(ctx context.Context, skip int64, limit int64) ([]*domain.Order, error) {
	return s.orderRepository.List(ctx, skip, limit)
}

func (s *OrderService) ListOrdersFiltered(ctx context.Context, status string, skip int64, limit int64) ([]*domain.Order, error) {
	if status == "" {
		return s.orderRepository.List(ctx, skip, limit)
	}
	return s.orderRepository.GetByStatus(ctx, status, skip, limit)
}

func (s *OrderService) CountOrders(ctx context.Context, status string) (int64, error) {
	return s.orderRepository.Count(ctx, status)
}

// SetPreference asocia un ID de preferencia de Mercado Pago a una orden existente.
// Llamado por el handler de checkout después de crear la preferencia en MP.
func (s *OrderService) SetPreference(ctx context.Context, id primitive.ObjectID, preferenceID string) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	order.PreferenceID = preferenceID
	order.UpdatedAt = time.Now()
	return s.orderRepository.Update(ctx, order)
}

// SetTracking guarda el número de seguimiento del envío de una orden.
func (s *OrderService) SetTracking(ctx context.Context, id primitive.ObjectID, trackingNumber string) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	order.TrackingNumber = trackingNumber
	order.UpdatedAt = time.Now()
	return s.orderRepository.Update(ctx, order)
}

// ConfirmPayment marca la orden como pagada y registra el ID de pago de MP.
// Idempotente: si la orden ya está en estado "paid", retorna nil.
func (s *OrderService) ConfirmPayment(ctx context.Context, orderID primitive.ObjectID, paymentID string) error {
	order, err := s.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	if order.Status == "paid" {
		return nil // ya procesado, respuesta idempotente
	}
	if order.Status == "cancelled" {
		// Re-decrement stock: the order was cancelled (stock was restored), but MP confirmed payment late.
		for _, item := range order.Items {
			if item.ItemType == "kit" {
				continue
			}
			var stockErr error
			if item.VariantSKU == "" {
				stockErr = s.productRepository.DecrementProductStock(ctx, item.ProductID, item.Quantity)
			} else {
				stockErr = s.productRepository.DecrementVariantStock(ctx, item.ProductID, item.VariantSKU, item.Quantity)
			}
			if stockErr != nil {
				log.Printf("ConfirmPayment: stock decrement failed for product %s sku=%q qty=%d: %v",
					item.ProductID.Hex(), item.VariantSKU, item.Quantity, stockErr)
			}
		}
	}
	order.Status = "paid"
	order.PaymentID = paymentID
	order.UpdatedAt = time.Now()
	return s.orderRepository.Update(ctx, order)
}

// Checkout descuenta stock atómicamente y crea la orden. Recibe ítems ya validados
// (con UnitPrice calculado) provenientes de CartService.ValidateCart.
func (s *OrderService) Checkout(ctx context.Context, userID primitive.ObjectID, items []domain.CartItem, shipping domain.ShippingDetails, customerName, customerEmail, customerPhone, dniCuit, notes, paymentMethod, deliveryMethod string, facturaA *domain.FacturaA, shippingCost, discount, couponDiscount float64, couponCode string) (*domain.Order, error) {
	if userID.IsZero() {
		return nil, errors.New("ID de usuario requerido")
	}
	if len(items) == 0 {
		return nil, errors.New("el carrito no puede estar vacío")
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	total := subtotal + shippingCost - discount - couponDiscount
	if total < 0 {
		total = 0
	}

	// Decrementar stock atómicamente; rastrear éxitos para rollback en caso de fallo parcial
	type decremented struct {
		productID primitive.ObjectID
		sku       string
		qty       int
	}
	var done []decremented

	for _, item := range items {
		if item.ItemType == "kit" {
			done = append(done, decremented{item.ProductID, item.VariantSKU, item.Quantity})
			continue
		}
		var decrementErr error
		if item.VariantSKU == "" {
			decrementErr = s.productRepository.DecrementProductStock(ctx, item.ProductID, item.Quantity)
		} else {
			decrementErr = s.productRepository.DecrementVariantStock(ctx, item.ProductID, item.VariantSKU, item.Quantity)
		}
		if decrementErr != nil {
			for _, d := range done {
				if d.sku == "" {
					_ = s.productRepository.IncrementProductStock(ctx, d.productID, d.qty)
				} else {
					_ = s.productRepository.IncrementVariantStock(ctx, d.productID, d.sku, d.qty)
				}
			}
			return nil, ErrInsufficientStock
		}
		done = append(done, decremented{item.ProductID, item.VariantSKU, item.Quantity})
	}

	now := time.Now()
	order := &domain.Order{
		UserID:          userID,
		Items:           items,
		Total:           total,
		Status:          "pending",
		CustomerName:    customerName,
		CustomerEmail:   customerEmail,
		CustomerPhone:   customerPhone,
		DniCuit:         dniCuit,
		Notes:           notes,
		PaymentMethod:   paymentMethod,
		DeliveryMethod:  deliveryMethod,
		ShippingDetails: shipping,
		ShippingCost:    shippingCost,
		Discount:        discount,
		CouponCode:      couponCode,
		CouponDiscount:  couponDiscount,
		FacturaA:        facturaA,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.orderRepository.Create(ctx, order); err != nil {
		for _, d := range done {
			if d.sku == "" {
				_ = s.productRepository.IncrementProductStock(ctx, d.productID, d.qty)
			} else {
				_ = s.productRepository.IncrementVariantStock(ctx, d.productID, d.sku, d.qty)
			}
		}
		return nil, err
	}

	return order, nil
}

// ExpireAbandonedOrders cancela órdenes "pending" más antiguas que `maxAge` y
// restaura el stock de sus items. Devuelve la cantidad de órdenes expiradas.
func (s *OrderService) ExpireAbandonedOrders(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)
	orders, err := s.orderRepository.FindPendingOlderThan(ctx, cutoff)
	if err != nil {
		return 0, err
	}

	var expired int
	for _, order := range orders {
		for _, item := range order.Items {
			if item.ItemType == "kit" {
				continue
			}
			if item.VariantSKU == "" {
				_ = s.productRepository.IncrementProductStock(ctx, item.ProductID, item.Quantity)
			} else {
				_ = s.productRepository.IncrementVariantStock(ctx, item.ProductID, item.VariantSKU, item.Quantity)
			}
		}
		if err := s.orderRepository.UpdateStatus(ctx, order.ID, "cancelled"); err == nil {
			expired++
		}
	}
	return expired, nil
}
