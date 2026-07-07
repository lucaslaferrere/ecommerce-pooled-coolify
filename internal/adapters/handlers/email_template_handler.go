package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailTemplateHandler struct {
	svc *services.EmailTemplateService
}

func NewEmailTemplateHandler(svc *services.EmailTemplateService) *EmailTemplateHandler {
	return &EmailTemplateHandler{svc: svc}
}

// GET /admin/email-templates
func (h *EmailTemplateHandler) List(c *gin.Context) {
	templates, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// POST /admin/email-templates
func (h *EmailTemplateHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req.Name, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// DELETE /admin/email-templates/:id
func (h *EmailTemplateHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plantilla eliminada"})
}
