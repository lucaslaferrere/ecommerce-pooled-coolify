package handlers

import (
	"log"
	"net/http"
	"strconv"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderHandler maneja las peticiones HTTP relacionadas con órdenes
type OrderHandler struct {
	orderService   *services.OrderService
	cartService    *services.CartService
	paymentService *services.PaymentService
}

// NewOrderHandler crea una nueva instancia de OrderHandler
func NewOrderHandler(
	orderService *services.OrderService,
	cartService *services.CartService,
	paymentService *services.PaymentService,
) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		cartService:    cartService,
		paymentService: paymentService,
	}
}

// checkoutRequest es el payload que envía el frontend al hacer checkout.
// Los items vienen del localStorage del carrito anónimo/autenticado.
type checkoutRequest struct {
	Items           []domain.CartItem      `json:"items" binding:"required,min=1"`
	ShippingDetails domain.ShippingDetails `json:"shipping_details" binding:"required"`
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

	sd := req.ShippingDetails
	if sd.Address == "" || sd.City == "" || sd.PostalCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipping_details requiere address, city y postal_code"})
		return
	}

	// Paso 1: validar carrito y calcular precios
	cart, err := h.cartService.ValidateCart(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Paso 2: descontar stock y persistir la orden (estado: "pending")
	order, err := h.orderService.Checkout(c.Request.Context(), userID, cart.Items, req.ShippingDetails)
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
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

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

// CancelOrder maneja PUT /api/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.orderService.CancelOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// ListAllOrders maneja GET /api/orders/admin/list (solo admin)
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	orders, err := h.orderService.ListOrders(c.Request.Context(), skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": len(orders)})
}
