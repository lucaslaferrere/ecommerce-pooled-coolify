package handlers

import (
	"net/http"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type WizardRecommendationHandler struct {
	service *services.WizardRecommendationService
}

func NewWizardRecommendationHandler(service *services.WizardRecommendationService) *WizardRecommendationHandler {
	return &WizardRecommendationHandler{service: service}
}

// CreateRecommendation maneja POST /api/v1/wizard-recommendations
func (h *WizardRecommendationHandler) CreateRecommendation(c *gin.Context) {
	var req struct {
		PoolSize         string `json:"pool_size"`
		PoolType         string `json:"pool_type"`
		UsageType        string `json:"usage_type"`
		ControlType      string `json:"control_type"`
		RecommendedKitID string `json:"recommended_kit_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec := &domain.WizardRecommendation{
		PoolSize:         req.PoolSize,
		PoolType:         req.PoolType,
		UsageType:        req.UsageType,
		ControlType:      req.ControlType,
		RecommendedKitID: req.RecommendedKitID,
		CreatedAt:        time.Now(),
	}

	if err := h.service.LogRecommendation(c.Request.Context(), rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "recomendación registrada"})
}
