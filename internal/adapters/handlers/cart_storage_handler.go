package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartStorageHandler struct {
	svc          *services.CartStorageService
	couponSvc    *services.CouponService
	emailService *services.EmailService
	userService  *services.UserService
	templateSvc  *services.EmailTemplateService
}

func NewCartStorageHandler(
	svc *services.CartStorageService,
	couponSvc *services.CouponService,
	emailService *services.EmailService,
	userService *services.UserService,
	templateSvc *services.EmailTemplateService,
) *CartStorageHandler {
	return &CartStorageHandler{
		svc:          svc,
		couponSvc:    couponSvc,
		emailService: emailService,
		userService:  userService,
		templateSvc:  templateSvc,
	}
}

func cartUserIDFromCtx(c *gin.Context) (primitive.ObjectID, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(raw.(string))
	if err != nil {
		return primitive.NilObjectID, false
	}
	return id, true
}

// PUT /cart — guarda el carrito del usuario logueado.
func (h *CartStorageHandler) SaveMyCart(c *gin.Context) {
	userID, ok := cartUserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	var req struct {
		Items []domain.CartItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Save(c.Request.Context(), userID, req.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "carrito guardado"})
}

// DELETE /cart — vacía el carrito del usuario logueado.
func (h *CartStorageHandler) ClearMyCart(c *gin.Context) {
	userID, ok := cartUserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	if err := h.svc.Clear(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "carrito vaciado"})
}

// GET /admin/carts — lista carritos no vacíos enriquecidos con email del usuario.
func (h *CartStorageHandler) ListCarts(c *gin.Context) {
	carts, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type cartRow struct {
		UserID             string            `json:"user_id"`
		Email              string            `json:"email"`
		Items              []domain.CartItem `json:"items"`
		ItemCount          int               `json:"item_count"`
		Total              float64           `json:"total"`
		UpdatedAt          interface{}       `json:"updated_at"`
		LastReminderSentAt interface{}       `json:"last_reminder_sent_at"`
	}

	rows := make([]cartRow, 0, len(carts))
	for _, cart := range carts {
		var total float64
		var count int
		for _, it := range cart.Items {
			total += it.UnitPrice * float64(it.Quantity)
			count += it.Quantity
		}
		email := ""
		if u, err := h.userService.GetUser(c.Request.Context(), cart.UserID); err == nil && u != nil {
			email = u.Email
		}
		var lastReminder interface{}
		if cart.LastReminderSentAt != nil {
			lastReminder = cart.LastReminderSentAt
		}
		rows = append(rows, cartRow{
			UserID:             cart.UserID.Hex(),
			Email:              email,
			Items:              cart.Items,
			ItemCount:          count,
			Total:              total,
			UpdatedAt:          cart.UpdatedAt,
			LastReminderSentAt: lastReminder,
		})
	}
	c.JSON(http.StatusOK, gin.H{"carts": rows})
}

// POST /admin/carts/:userID/send-coupon — envía el email de recuperación.
func (h *CartStorageHandler) SendCoupon(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var req struct {
		TemplateID string `json:"template_id" binding:"required"`
		CouponID   string `json:"coupon_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "el carrito está vacío o no existe"})
		return
	}
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	templateID, err := primitive.ObjectIDFromHex(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plantilla inválida"})
		return
	}
	template, err := h.templateSvc.GetByID(c.Request.Context(), templateID)
	if err != nil || template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plantilla no encontrada"})
		return
	}

	// El cupón es opcional.
	var coupon *domain.Coupon
	if req.CouponID != "" {
		couponID, err := primitive.ObjectIDFromHex(req.CouponID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cupón inválido"})
			return
		}
		coupon, err = h.couponSvc.GetByID(c.Request.Context(), couponID)
		if err != nil || coupon == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "cupón no encontrado"})
			return
		}
	}

	var total float64
	for _, it := range cart.Items {
		total += it.UnitPrice * float64(it.Quantity)
	}

	if err := h.emailService.SendAbandonedCartEmail(user.Email, template.Body, coupon, cart.Items, total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo enviar el email: " + err.Error()})
		return
	}
	_ = h.svc.MarkReminderSent(c.Request.Context(), userID) // no crítico
	c.JSON(http.StatusOK, gin.H{"message": "email enviado a " + user.Email})
}
