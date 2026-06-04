package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// EmailService envía emails transaccionales usando la API de Resend
type EmailService struct {
	apiKey string
	from   string
}

// NewEmailService crea una instancia de EmailService.
// Si apiKey está vacío el servicio devuelve error en cada envío.
func NewEmailService(apiKey, from string) *EmailService {
	return &EmailService{apiKey: apiKey, from: from}
}

type resendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// Send envía un email HTML al destinatario indicado
func (s *EmailService) Send(to, subject, html string) error {
	if s.apiKey == "" {
		return fmt.Errorf("servicio de email no configurado")
	}

	payload := resendEmailPayload{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
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

// SendPasswordResetCode envía el email con el código de recuperación al usuario
func (s *EmailService) SendPasswordResetCode(to, code string) error {
	subject := "Recuperación de contraseña"
	html := buildPasswordResetHTML(to, code)
	return s.Send(to, subject, html)
}

func buildPasswordResetHTML(email, code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8" /></head>
<body style="font-family:sans-serif;background:#f9fafb;padding:32px;margin:0;">
  <div style="max-width:480px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
    <div style="background:#0B1F3A;padding:24px 32px;">
      <h1 style="color:#ffffff;margin:0;font-size:20px;">Recuperar contraseña</h1>
      <p style="color:#94a3b8;margin:6px 0 0;font-size:14px;">pooled.com.ar</p>
    </div>
    <div style="padding:32px;">
      <p style="color:#374151;font-size:15px;margin:0 0 24px;">
        Recibimos una solicitud para restablecer la contraseña de <strong>%s</strong>.
      </p>
      <p style="color:#374151;font-size:15px;margin:0 0 12px;">Tu código de verificación es:</p>
      <div style="text-align:center;margin:0 0 28px;">
        <span style="display:inline-block;background:#f3f4f6;border-radius:10px;padding:16px 40px;font-size:36px;font-weight:700;letter-spacing:10px;color:#0B1F3A;font-family:monospace;">%s</span>
      </div>
      <p style="color:#6b7280;font-size:13px;margin:0 0 8px;">
        Este código es válido por <strong>15 minutos</strong>.
      </p>
      <p style="color:#6b7280;font-size:13px;margin:0;">
        Si no solicitaste este cambio podés ignorar este mensaje; tu contraseña no se modificará.
      </p>
    </div>
    <div style="background:#f9fafb;padding:16px 32px;border-top:1px solid #e5e7eb;">
      <p style="color:#9ca3af;font-size:12px;margin:0;text-align:center;">pooled.com.ar — Este es un mensaje automático, no respondas este email.</p>
    </div>
  </div>
</body>
</html>`, email, code)
}
