package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	svc *services.SettingService
}

func NewSettingHandler(svc *services.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

// GET /admin/settings/abandoned-cart-email
func (h *SettingHandler) GetAbandonedCartEmail(c *gin.Context) {
	body, err := h.svc.GetAbandonedCartBody(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"body":         body,
		"default_body": h.svc.DefaultAbandonedCartBody(),
	})
}

// PUT /admin/settings/abandoned-cart-email
func (h *SettingHandler) SetAbandonedCartEmail(c *gin.Context) {
	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetAbandonedCartBody(c.Request.Context(), req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plantilla guardada"})
}
