package handlers

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxInvoiceBytes = 8 * 1024 * 1024 // 8 MB

type InvoiceHandler struct {
	svc *services.InvoiceService
}

func NewInvoiceHandler(svc *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// POST /admin/orders/:id/invoice — multipart con campo "file" (PDF).
func (h *InvoiceHandler) Upload(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo"})
		return
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el archivo debe ser un PDF"})
		return
	}
	if fileHeader.Size > maxInvoiceBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el archivo supera los 8 MB"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el archivo"})
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el archivo"})
		return
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(raw)

	sentAt, err := h.svc.SaveAndSend(c.Request.Context(), orderID, fileHeader.Filename, pdfBase64)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pedido no encontrado"})
			return
		}
		// La factura pudo guardarse aunque el envío falle.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "la factura se guardó pero no se pudo enviar: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoice_sent_at": sentAt})
}

// POST /admin/orders/:id/invoice/resend — reenvía la factura guardada.
func (h *InvoiceHandler) Resend(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	sentAt, err := h.svc.Resend(c.Request.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvoiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "este pedido no tiene factura cargada"})
		case errors.Is(err, services.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "pedido no encontrado"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo reenviar: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoice_sent_at": sentAt})
}
