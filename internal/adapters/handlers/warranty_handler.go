package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WarrantyHandler struct {
	resendKey  string
	resendFrom string
}

func NewWarrantyHandler(resendKey, resendFrom string) *WarrantyHandler {
	return &WarrantyHandler{resendKey: resendKey, resendFrom: resendFrom}
}

type warrantyRequest struct {
	Serial       string `json:"serial"        binding:"required"`
	Model        string `json:"model"         binding:"required"`
	ProblemDesc  string `json:"problem_desc"  binding:"required"`
	Invoice      string `json:"invoice"       binding:"required"`
	PurchaseDate string `json:"purchase_date" binding:"required"`
	Name         string `json:"name"          binding:"required"`
	Email        string `json:"email"         binding:"required"`
	Phone        string `json:"phone"         binding:"required"`
	Message      string `json:"message"`
}

func (h *WarrantyHandler) Submit(c *gin.Context) {
	var req warrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos incompletos"})
		return
	}

	if h.resendKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de email no configurado"})
		return
	}

	subject := fmt.Sprintf("Solicitud de Garantía — %s — %s", req.Model, req.Name)
	if err := h.sendEmail(req.Email, subject, buildWarrantyHTML(req)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo enviar el email. Intentá más tarde."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── Resend REST call ──────────────────────────────────────────────────────────

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo []string `json:"reply_to,omitempty"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (h *WarrantyHandler) sendEmail(replyTo, subject, html string) error {
	payload := resendPayload{
		From:    h.resendFrom,
		To:      []string{"garantia@pooled.com.ar"},
		Subject: subject,
		HTML:    html,
	}
	if replyTo != "" {
		payload.ReplyTo = []string{replyTo}
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.resendKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend error: status %d", resp.StatusCode)
	}
	return nil
}

// ── HTML template ─────────────────────────────────────────────────────────────

func buildWarrantyHTML(r warrantyRequest) string {
	extra := ""
	if r.Message != "" {
		extra = fmt.Sprintf(`
		<tr><td colspan="2" style="padding-top:16px;font-weight:600;color:#0B1F3A;">Mensaje adicional</td></tr>
		<tr><td colspan="2" style="padding:8px 0;color:#374151;">%s</td></tr>`, r.Message)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8" /></head>
<body style="font-family:sans-serif;background:#f9fafb;padding:32px;">
  <div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <div style="background:#0B1F3A;padding:24px 32px;">
      <h1 style="color:#ffffff;margin:0;font-size:20px;">🛡️ Solicitud de Garantía</h1>
      <p style="color:#94a3b8;margin:4px 0 0;font-size:14px;">pooled.com.ar</p>
    </div>
    <div style="padding:32px;">
      <table style="width:100%%;border-collapse:collapse;font-size:14px;">
        <tr><td colspan="2" style="padding-bottom:8px;font-weight:600;color:#0B1F3A;border-bottom:1px solid #e5e7eb;margin-bottom:12px;">Producto</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;width:160px;">Número de serie</td><td style="color:#111827;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Modelo / Línea</td><td style="color:#111827;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Problema</td><td style="color:#111827;">%s</td></tr>

        <tr><td colspan="2" style="padding:16px 0 8px;font-weight:600;color:#0B1F3A;border-bottom:1px solid #e5e7eb;">Compra</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Factura / Comprobante</td><td style="color:#111827;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Fecha de compra</td><td style="color:#111827;">%s</td></tr>

        <tr><td colspan="2" style="padding:16px 0 8px;font-weight:600;color:#0B1F3A;border-bottom:1px solid #e5e7eb;">Contacto</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Nombre</td><td style="color:#111827;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Email</td><td style="color:#111827;">%s</td></tr>
        <tr><td style="padding:6px 0;color:#6b7280;">Teléfono</td><td style="color:#111827;">%s</td></tr>
        %s
      </table>
      <div style="margin-top:24px;padding:14px;background:#fef3c7;border-radius:8px;font-size:13px;color:#92400e;">
        ⚠️ Recordá pedirle al cliente que adjunte foto del producto y video del problema respondiendo este email.
      </div>
    </div>
  </div>
</body>
</html>`,
		r.Serial, r.Model, r.ProblemDesc,
		r.Invoice, r.PurchaseDate,
		r.Name, r.Email, r.Phone,
		extra,
	)
}
