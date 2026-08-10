package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mutatingMethods son los únicos métodos que se auditan. Los GET (lecturas)
// no se registran a propósito — el objetivo es "quién cambió qué", no
// "quién miró qué".
var mutatingMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

// sensitiveBodyKeys nunca se guardan en el detalle del log, aunque vengan en
// el body de la request.
var sensitiveBodyKeys = map[string]bool{
	"password": true, "new_password": true, "current_password": true,
	"token": true, "access_token": true, "refresh_token": true,
}

const maxAuditDetailsLen = 500

// AuditLogMiddleware registra cada acción de escritura (POST/PUT/PATCH/DELETE)
// hecha por un admin autenticado. Va DESPUÉS de AuthMiddleware+AdminMiddleware
// en la cadena, así ya tiene user_id/email en el contexto. Nunca bloquea ni
// hace fallar la request: si algo del logging falla, solo se loguea en server.
func AuditLogMiddleware(auditService *services.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mutatingMethods[c.Request.Method] {
			c.Next()
			return
		}

		details := readAuditDetails(c)

		c.Next()

		adminIDRaw, _ := c.Get("user_id")
		adminIDStr, _ := adminIDRaw.(string)
		adminID, err := primitive.ObjectIDFromHex(adminIDStr)
		if err != nil {
			return // no autenticado (no debería pasar en rutas admin, pero no rompemos nada)
		}
		emailRaw, _ := c.Get("email")
		adminEmail, _ := emailRaw.(string)

		route := c.FullPath()
		path := c.Request.URL.Path
		status := c.Writer.Status()
		method := c.Request.Method

		go func() {
			if err := auditService.LogAction(context.Background(), adminID, adminEmail, method, route, path, details, status); err != nil {
				log.Printf("audit log: no se pudo guardar la entrada (%s %s): %v", method, route, err)
			}
		}()
	}
}

// readAuditDetails lee y restaura el body si es JSON (multipart/form-data se
// salta: puede traer archivos binarios pesados y no vale la pena capturarlo).
// Recorta a maxAuditDetailsLen y redacta claves sensibles.
func readAuditDetails(c *gin.Context) string {
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 8192))
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	for k := range parsed {
		if sensitiveBodyKeys[strings.ToLower(k)] {
			parsed[k] = "***"
		}
	}
	redacted, err := json.Marshal(parsed)
	if err != nil {
		return ""
	}
	s := string(redacted)
	if len(s) > maxAuditDetailsLen {
		s = s[:maxAuditDetailsLen] + "…"
	}
	return s
}
