package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// TrackEvent maneja POST /api/v1/events (público, fire-and-forget)
func (h *EventHandler) TrackEvent(c *gin.Context) {
	var req struct {
		Type      string                 `json:"type"`
		SessionID string                 `json:"session_id"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type requerido"})
		return
	}

	event := &domain.Event{
		Type:      req.Type,
		SessionID: req.SessionID,
		Payload:   req.Payload,
	}

	_ = h.service.Track(c.Request.Context(), event)
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// GetAnalytics maneja GET /api/v1/admin/analytics?period=7d (solo admin)
func (h *EventHandler) GetAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")
	if period != "7d" && period != "30d" {
		period = "7d"
	}

	report, err := h.service.GetAnalytics(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
