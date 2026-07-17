package handlers

import (
	"encoding/json"
	"net/http"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"
	"ecommerce-pooled/internal/realtime"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service  *services.EventService
	geo      *services.GeoService
	sessions *realtime.ActiveSessions
	hub      *realtime.Hub
}

func NewEventHandler(service *services.EventService, geo *services.GeoService, sessions *realtime.ActiveSessions, hub *realtime.Hub) *EventHandler {
	return &EventHandler{service: service, geo: geo, sessions: sessions, hub: hub}
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
	// IP y geo se completan server-side, nunca desde el frontend.
	event.IP = c.ClientIP()
	event.Country, event.City, event.Province, event.Lat, event.Lng = h.geo.Lookup(event.IP)
	h.sessions.Touch(req.SessionID)

	_ = h.service.Track(c.Request.Context(), event)

	msg, _ := json.Marshal(map[string]any{
		"type":       "event",
		"event_type": event.Type,
		"session_id": event.SessionID,
	})
	h.hub.Broadcast(string(msg))
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// GetTraffic maneja GET /api/v1/admin/traffic (solo admin)
func (h *EventHandler) GetTraffic(c *gin.Context) {
	report, err := h.service.GetTrafficReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetAnalytics maneja GET /api/v1/admin/analytics?period=7d (solo admin)
func (h *EventHandler) GetAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")
	switch period {
	case "1d", "7d", "30d", "1y":
		// valid
	default:
		period = "7d"
	}

	report, err := h.service.GetAnalytics(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
