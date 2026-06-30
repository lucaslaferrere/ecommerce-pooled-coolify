package handlers

import (
	"errors"
	"net/http"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CouponHandler struct {
	svc *services.CouponService
}

func NewCouponHandler(svc *services.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

// POST /coupons/validate — public endpoint called from checkout
func (h *CouponHandler) Validate(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código requerido"})
		return
	}

	// Validación pública (sin sesión): solo verifica el límite total.
	// El límite por cliente se hace cumplir en el checkout.
	coupon, err := h.svc.ValidateForUser(c.Request.Context(), req.Code, primitive.NilObjectID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCouponNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Cupón no encontrado"})
		case errors.Is(err, services.ErrCouponInactive):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Este cupón no está activo"})
		case errors.Is(err, services.ErrCouponExpired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Este cupón está vencido"})
		case errors.Is(err, services.ErrCouponMaxUses):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Este cupón ya alcanzó su límite de usos"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":             coupon.Code,
		"discount_percent": coupon.DiscountPercent,
	})
}

// GET /admin/coupons
func (h *CouponHandler) List(c *gin.Context) {
	coupons, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

// POST /admin/coupons
func (h *CouponHandler) Create(c *gin.Context) {
	var req struct {
		Code            string     `json:"code"             binding:"required"`
		DiscountPercent float64    `json:"discount_percent" binding:"required,gt=0,lte=100"`
		Active          bool       `json:"active"`
		MaxUses         int        `json:"max_uses"`
		MaxUsesPerUser  int        `json:"max_uses_per_user"`
		ExpiresAt       *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon := &domain.Coupon{
		Code:            req.Code,
		DiscountPercent: req.DiscountPercent,
		Active:          req.Active,
		MaxUses:         req.MaxUses,
		MaxUsesPerUser:  req.MaxUsesPerUser,
		ExpiresAt:       req.ExpiresAt,
	}
	if err := h.svc.Create(c.Request.Context(), coupon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, coupon)
}

// PUT /admin/coupons/:id
func (h *CouponHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Code            string     `json:"code"`
		DiscountPercent float64    `json:"discount_percent"`
		Active          bool       `json:"active"`
		MaxUses         int        `json:"max_uses"`
		MaxUsesPerUser  int        `json:"max_uses_per_user"`
		ExpiresAt       *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon := &domain.Coupon{
		ID:              id,
		Code:            req.Code,
		DiscountPercent: req.DiscountPercent,
		Active:          req.Active,
		MaxUses:         req.MaxUses,
		MaxUsesPerUser:  req.MaxUsesPerUser,
		ExpiresAt:       req.ExpiresAt,
	}
	if err := h.svc.Update(c.Request.Context(), coupon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, coupon)
}

// DELETE /admin/coupons/:id
func (h *CouponHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cupón eliminado"})
}
