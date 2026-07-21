package handlers

import (
	"encoding/json"
	"net/http"
	"time"

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

// GetSalesSeries maneja GET /api/v1/admin/analytics/sales (solo admin).
// Query: from, to (YYYY-MM-DD o RFC3339) y granularity=month|day.
// Por defecto: últimos 12 meses con granularidad mensual.
func (h *EventHandler) GetSalesSeries(c *gin.Context) {
	granularity := c.DefaultQuery("granularity", "month")
	if granularity != "day" && granularity != "month" {
		granularity = "month"
	}

	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	parse := func(v string, def time.Time) time.Time {
		if v == "" {
			return def
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		if t, err := time.ParseInLocation("2006-01-02", v, loc); err == nil {
			return t
		}
		return def
	}

	// Default: desde el 1° de mes, 11 meses atrás (12 meses incluyendo el actual).
	defFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -11, 0)
	from := parse(c.Query("from"), defFrom)
	to := parse(c.Query("to"), now)

	series, err := h.service.GetSalesSeries(c.Request.Context(), from, to, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"series": series, "granularity": granularity})
}

// GetSalesOverview maneja GET /api/v1/admin/analytics/overview (solo admin).
// Query: from, to (YYYY-MM-DD o RFC3339) y granularity=month|day.
// El período previo se calcula con el mismo span inmediatamente anterior.
func (h *EventHandler) GetSalesOverview(c *gin.Context) {
	granularity := c.DefaultQuery("granularity", "day")
	if granularity != "day" && granularity != "month" {
		granularity = "day"
	}

	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	parse := func(v string, def time.Time) time.Time {
		if v == "" {
			return def
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		if t, err := time.ParseInLocation("2006-01-02", v, loc); err == nil {
			return t
		}
		return def
	}

	from := parse(c.Query("from"), now.AddDate(0, 0, -29))
	to := parse(c.Query("to"), now)

	// Período previo: mismo span, inmediatamente anterior.
	span := to.Sub(from)
	prevTo := from.Add(-time.Millisecond)
	prevFrom := prevTo.Add(-span)

	overview, err := h.service.GetSalesOverview(c.Request.Context(), from, to, prevFrom, prevTo, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
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
