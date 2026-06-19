package handlers

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ── Shipping & discount calculations (server-side; mirrors shipping.ts) ──────

const (
	shippingBaseAMBA         = 9_000.0
	shippingFreeThreshold    = 1_000_000.0
	transferDiscountRate     = 0.035
)

var provinceZones = map[string]string{
	"CABA": "Z1", "Buenos Aires (GBA)": "Z1",
	"Buenos Aires (interior)": "Z2", "La Pampa": "Z2",
	"Córdoba": "Z3", "Santa Fe": "Z3", "Entre Ríos": "Z3",
	"Mendoza": "Z4", "San Juan": "Z4", "San Luis": "Z4",
	"Tucumán": "Z5", "Salta": "Z5", "Jujuy": "Z5",
	"Santiago del Estero": "Z5", "Catamarca": "Z5", "La Rioja": "Z5",
	"Misiones": "Z5", "Corrientes": "Z5", "Chaco": "Z5", "Formosa": "Z5",
	"Neuquén": "Z6", "Río Negro": "Z6", "Chubut": "Z6", "Santa Cruz": "Z6",
	"Tierra del Fuego": "Z7",
}

var zoneMultipliers = map[string]float64{
	"Z1": 1.00, "Z2": 1.15, "Z3": 1.30, "Z4": 1.55,
	"Z5": 1.80, "Z6": 2.20, "Z7": 2.90,
}

// serverCalcShipping replicates the frontend calcShipping(province, 1kg, domicilio, subtotal).
func serverCalcShipping(province, deliveryMethod string, subtotal float64) float64 {
	if deliveryMethod == "retirar" || province == "" {
		return 0
	}
	zoneKey, ok := provinceZones[province]
	if !ok {
		return 0
	}
	if subtotal >= shippingFreeThreshold {
		return 0
	}
	return math.Round(shippingBaseAMBA * zoneMultipliers[zoneKey])
}

// serverCalcDiscount computes the transfer discount server-side.
func serverCalcDiscount(subtotal float64, paymentMethod string) float64 {
	if paymentMethod == "transferencia" {
		return math.Round(subtotal * transferDiscountRate)
	}
	return 0
}

// OrderHandler maneja las peticiones HTTP relacionadas con órdenes
type OrderHandler struct {
	orderService   *services.OrderService
	cartService    *services.CartService
	paymentService *services.PaymentService
	emailService   *services.EmailService
}

// NewOrderHandler crea una nueva instancia de OrderHandler
func NewOrderHandler(
	orderService *services.OrderService,
	cartService *services.CartService,
	paymentService *services.PaymentService,
	emailService *services.EmailService,
) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		cartService:    cartService,
		paymentService: paymentService,
		emailService:   emailService,
	}
}

// FacturaARequest contiene los datos de Factura A opcionales
type FacturaARequest struct {
	RazonSocial string `json:"razon_social"`
	CUIT        string `json:"cuit"`
}

// checkoutRequest es el payload que envía el frontend al hacer checkout.
// Nota: shipping_cost y discount NO se aceptan del cliente — se calculan server-side.
type checkoutRequest struct {
	Items            []domain.CartItem `json:"items" binding:"required,min=1"`
	CustomerName     string            `json:"customer_name" binding:"required"`
	CustomerEmail    string            `json:"customer_email" binding:"required,email"`
	CustomerPhone    string            `json:"customer_phone" binding:"required"`
	DniCuit          string            `json:"dni_cuit"`
	ShippingAddress  string            `json:"shipping_address"`
	ShippingCity     string            `json:"shipping_city"`
	ShippingProvince string            `json:"shipping_province"`
	ShippingZip      string            `json:"shipping_zip"`
	Notes            string            `json:"notes"`
	PaymentMethod    string            `json:"payment_method"`
	DeliveryMethod   string            `json:"delivery_method"`
	FacturaA         *FacturaARequest  `json:"factura_a"`
}

// Checkout maneja POST /api/v1/checkout
// Valida el carrito, descuenta stock y crea la orden en un solo flujo.
func (h *OrderHandler) Checkout(c *gin.Context) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}

	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sd := domain.ShippingDetails{
		Address:    req.ShippingAddress,
		City:       req.ShippingCity,
		Province:   req.ShippingProvince,
		PostalCode: req.ShippingZip,
	}

	var facturaA *domain.FacturaA
	if req.FacturaA != nil && req.FacturaA.RazonSocial != "" {
		facturaA = &domain.FacturaA{
			RazonSocial: req.FacturaA.RazonSocial,
			CUIT:        req.FacturaA.CUIT,
		}
	}

	// Paso 1: validar carrito y calcular precios
	cart, err := h.cartService.ValidateCart(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Calcular envío y descuento server-side (nunca confiar en valores del cliente)
	shippingCost := serverCalcShipping(req.ShippingProvince, req.DeliveryMethod, cart.Total)
	discount := serverCalcDiscount(cart.Total, req.PaymentMethod)

	// Paso 2: descontar stock y persistir la orden (estado: "pending")
	order, err := h.orderService.Checkout(c.Request.Context(), userID, cart.Items, sd, req.CustomerName, req.CustomerEmail, req.CustomerPhone, req.DniCuit, req.Notes, req.PaymentMethod, req.DeliveryMethod, facturaA, shippingCost, discount)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Paso 3: crear preferencia de pago en Mercado Pago
	// Si el servicio no está configurado o falla, el checkout igual retorna 201
	// con init_point vacío; el frontend debe manejar esa situación.
	pref, mpErr := h.paymentService.CreatePreference(c.Request.Context(), order)
	if mpErr != nil {
		log.Printf("checkout: error creando preferencia MP para orden %s: %v", order.ID.Hex(), mpErr)
		pref = &services.PaymentPreference{}
	}

	// Paso 4: persistir el preference_id en la orden (best-effort)
	if pref.PreferenceID != "" {
		if err := h.orderService.SetPreference(c.Request.Context(), order.ID, pref.PreferenceID); err != nil {
			log.Printf("checkout: error guardando preference_id en orden %s: %v", order.ID.Hex(), err)
		} else {
			order.PreferenceID = pref.PreferenceID
		}
	}

	// Notificar a los dueños del nuevo pedido (fire-and-forget)
	go h.emailService.SendOwnerNewOrder(order)

	c.JSON(http.StatusCreated, gin.H{
		"order":              order,
		"init_point":         pref.InitPoint,
		"sandbox_init_point": pref.SandboxInitPoint,
	})
}

// GetMyOrders maneja GET /api/v1/orders/me
// Retorna las órdenes del usuario autenticado.
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}

	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": len(orders)})
}

// UpdateAdminOrderStatus maneja PATCH /api/v1/admin/orders/:id/status
func (h *OrderHandler) UpdateAdminOrderStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Status         string `json:"status" binding:"required"`
		TrackingNumber string `json:"tracking_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Guardar tracking y avisar al cliente cuando el pedido se despacha
	if req.Status == "shipped" && req.TrackingNumber != "" {
		if err := h.orderService.SetTracking(c.Request.Context(), id, req.TrackingNumber); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "estado actualizado pero no se pudo guardar el tracking: " + err.Error()})
			return
		}
		go func() {
			order, err := h.orderService.GetOrder(context.Background(), id)
			if err != nil {
				log.Printf("email envío: error obteniendo orden %s: %v", id.Hex(), err)
				return
			}
			if err := h.emailService.SendShippingNotification(order); err != nil {
				log.Printf("email envío orden %s: %v", id.Hex(), err)
			} else {
				log.Printf("email envío enviado a %s (orden %s, tracking %s)", order.CustomerEmail, id.Hex(), order.TrackingNumber)
			}
		}()
	}

	// Emails post-cambio de estado (fire-and-forget)
	go func() {
		order, err := h.orderService.GetOrder(context.Background(), id)
		if err != nil {
			log.Printf("post-status email: error obteniendo orden %s: %v", id.Hex(), err)
			return
		}
		// Confirmación al cliente cuando pasa a "paid"
		if req.Status == "paid" {
			if err := h.emailService.SendOrderConfirmation(order); err != nil {
				log.Printf("email confirmación orden %s: %v", id.Hex(), err)
			}
		}
		// Notificación de envío al cliente cuando pasa a "shipped"
		if req.Status == "shipped" {
			if err := h.emailService.SendShippingNotification(order); err != nil {
				log.Printf("email envío orden %s: %v", id.Hex(), err)
			}
		}
		// Notificación a los dueños en cualquier cambio de estado
		h.emailService.SendOwnerStatusChange(order)
	}()

	c.JSON(http.StatusOK, gin.H{"message": "estado actualizado"})
}

// ──────────────────────────────────────────────────────────────────────────────
// Handlers legacy (rutas /api/orders existentes)
// ──────────────────────────────────────────────────────────────────────────────

// CreateOrder maneja POST /api/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order domain.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.orderService.CreateOrder(c.Request.Context(), &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// GetOrder maneja GET /api/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetUserOrders maneja GET /api/users/:userID/orders
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": len(orders)})
}

// UpdateOrderStatus maneja PUT /api/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "estado actualizado"})
}

// CancelOrder maneja PATCH /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Verify ownership before cancelling
	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "orden no encontrada"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "orden cancelada"})
}

// DeleteOrder maneja DELETE /api/orders/:id
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.orderService.DeleteOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "orden eliminada"})
}

// ListAllOrders maneja GET /api/v1/admin/orders
// Query params: page (1-based), limit (default 20), status (optional)
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	skip := (page - 1) * limit

	orders, err := h.orderService.ListOrdersFiltered(c.Request.Context(), status, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total, err := h.orderService.CountOrders(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}
