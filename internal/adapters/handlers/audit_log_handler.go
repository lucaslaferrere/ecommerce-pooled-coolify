package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLogHandler struct {
	svc *services.AuditLogService
}

func NewAuditLogHandler(svc *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

// List maneja GET /admin/audit-logs?page&limit&admin_id&from&to (solo admin).
// from/to en formato YYYY-MM-DD.
func (h *AuditLogHandler) List(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	if limit < 1 || limit > 200 {
		limit = 50
	}

	var adminID primitive.ObjectID
	if raw := c.Query("admin_id"); raw != "" {
		id, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "admin_id inválido"})
			return
		}
		adminID = id
	}

	var from, to time.Time
	if raw := c.Query("from"); raw != "" {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from inválido, usar YYYY-MM-DD"})
			return
		}
		from = t
	}
	if raw := c.Query("to"); raw != "" {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to inválido, usar YYYY-MM-DD"})
			return
		}
		to = t.Add(24*time.Hour - time.Millisecond) // incluye todo el día
	}

	logs, total, err := h.svc.List(c.Request.Context(), adminID, from, to, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ListAdmins maneja GET /admin/audit-logs/admins (solo admin) — para poblar
// el filtro del frontend con los admins que tienen actividad registrada.
func (h *AuditLogHandler) ListAdmins(c *gin.Context) {
	admins, err := h.svc.ListAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"admins": admins})
}
