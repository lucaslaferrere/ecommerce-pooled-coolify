package handlers

import (
	"io"
	"net/http"

	"ecommerce-pooled/internal/core/services"
	"ecommerce-pooled/internal/realtime"

	"github.com/gin-gonic/gin"
)

type RealtimeHandler struct {
	hub      *realtime.Hub
	events   *services.EventService
	sessions *realtime.ActiveSessions
}

func NewRealtimeHandler(hub *realtime.Hub, events *services.EventService, sessions *realtime.ActiveSessions) *RealtimeHandler {
	return &RealtimeHandler{hub: hub, events: events, sessions: sessions}
}

// Snapshot maneja GET /api/v1/admin/realtime/snapshot (solo admin).
func (h *RealtimeHandler) Snapshot(c *gin.Context) {
	snap, err := h.events.GetRealtimeSnapshot(c.Request.Context(), h.sessions.GetActiveSessionCount())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snap)
}

// Stream maneja GET /api/v1/admin/realtime/stream (SSE, solo admin).
func (h *RealtimeHandler) Stream(c *gin.Context) {
	ch := h.hub.Subscribe()
	defer h.hub.Unsubscribe(ch)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
