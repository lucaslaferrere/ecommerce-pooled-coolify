package handlers

import (
	"net/http"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type DistributorLeadHandler struct {
	service *services.DistributorLeadService
}

func NewDistributorLeadHandler(service *services.DistributorLeadService) *DistributorLeadHandler {
	return &DistributorLeadHandler{service: service}
}

// CreateLead maneja POST /api/v1/distributor-leads
func (h *DistributorLeadHandler) CreateLead(c *gin.Context) {
	var req struct {
		FullName string `json:"full_name" binding:"required"`
		Company  string `json:"company" binding:"required"`
		City     string `json:"city" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Message  string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead := &domain.DistributorLead{
		FullName:  req.FullName,
		Company:   req.Company,
		City:      req.City,
		Phone:     req.Phone,
		Email:     req.Email,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}

	if err := h.service.CreateLead(c.Request.Context(), lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "solicitud recibida"})
}
