package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"ecommerce-pooled/internal/core/domain"
)

// EmailService envía emails transaccionales usando la API de Resend.
type EmailService struct {
	apiKey      string
	from        string
	OwnerEmails []string
}

func NewEmailService(apiKey, from string, ownerEmails []string) *EmailService {
	return &EmailService{apiKey: apiKey, from: from, OwnerEmails: ownerEmails}
}

func (s *EmailService) Enabled() bool {
	return s.apiKey != ""
}

// Send envía un email HTML al destinatario indicado.
func (s *EmailService) Send(to, subject, html string) error {
	if s.apiKey == "" {
		return fmt.Errorf("servicio de email no configurado")
	}

	payload := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		HTML    string   `json:"html"`
	}{
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

// SendAbandonedCartEmail envía el email de recuperación de carrito con el cupón.
// bodyText es el template editable; se interpolan las variables y se envuelve en HTML.
func (s *EmailService) SendAbandonedCartEmail(toEmail, bodyText string, coupon *domain.Coupon, items []domain.CartItem, total float64) error {
	if !s.Enabled() {
		return nil
	}

	// Construir el listado de productos en texto.
	var productsText strings.Builder
	for _, it := range items {
		name := it.Name
		if name == "" {
			name = it.VariantSKU
		}
		fmt.Fprintf(&productsText, "- %s x%d\n", name, it.Quantity)
	}

	vencimiento := "sin vencimiento"
	if coupon.ExpiresAt != nil {
		vencimiento = coupon.ExpiresAt.Format("02/01/2006")
	}
	descuento := fmt.Sprintf("%g%%", coupon.DiscountPercent)
	totalStr := fmt.Sprintf("$%.0f", total)

	replacer := strings.NewReplacer(
		"{{codigo}}", coupon.Code,
		"{{descuento}}", descuento,
		"{{vencimiento}}", vencimiento,
		"{{productos}}", productsText.String(),
		"{{total}}", totalStr,
	)
	filled := replacer.Replace(bodyText)

	// Convertir saltos de línea a <br> y envolver en HTML simple.
	htmlBody := strings.ReplaceAll(filled, "\n", "<br>")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="es"><head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:-apple-system,Segoe UI,Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
        <tr><td style="background:#0ea5e9;padding:24px 32px;">
          <h1 style="margin:0;color:#fff;font-size:18px;">🛒 Tu carrito te espera</h1>
        </td></tr>
        <tr><td style="padding:28px 32px;color:#374151;font-size:14px;line-height:1.6;">
          %s
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`, htmlBody)

	subject := fmt.Sprintf("Te dejamos un %s de descuento 🎁", descuento)
	return s.Send(toEmail, subject, html)
}

// SendPasswordResetCode envía el código de recuperación de contraseña.
func (s *EmailService) SendPasswordResetCode(to, code string) error {
	return s.Send(to, "Recuperación de contraseña", buildPasswordResetHTML(to, code))
}

// SendOrderConfirmation manda el email de confirmación de compra al cliente.
func (s *EmailService) SendOrderConfirmation(order *domain.Order) error {
	if !s.Enabled() {
		return nil
	}
	ref := strings.ToUpper(order.ID.Hex()[len(order.ID.Hex())-8:])
	subject := fmt.Sprintf("✅ Pedido confirmado #%s — Pooled", ref)
	return s.Send(order.CustomerEmail, subject, buildOrderConfirmationHTML(order))
}

// ── HTML: password reset ──────────────────────────────────────────────────────

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
      <p style="color:#6b7280;font-size:13px;margin:0 0 8px;">Este código es válido por <strong>15 minutos</strong>.</p>
      <p style="color:#6b7280;font-size:13px;margin:0;">Si no solicitaste este cambio podés ignorar este mensaje.</p>
    </div>
    <div style="background:#f9fafb;padding:16px 32px;border-top:1px solid #e5e7eb;">
      <p style="color:#9ca3af;font-size:12px;margin:0;text-align:center;">pooled.com.ar — Este es un mensaje automático, no respondas este email.</p>
    </div>
  </div>
</body>
</html>`, email, code)
}

// ── HTML: order confirmation ──────────────────────────────────────────────────

func buildOrderConfirmationHTML(o *domain.Order) string {
	orderRef := strings.ToUpper(o.ID.Hex()[len(o.ID.Hex())-8:])

	var itemRows string
	for _, item := range o.Items {
		name := item.Name
		if name == "" {
			name = item.VariantSKU
		}
		itemRows += fmt.Sprintf(`
		<tr>
			<td style="padding:10px 0;border-bottom:1px solid #f3f4f6;color:#111827;">%s</td>
			<td style="padding:10px 0;border-bottom:1px solid #f3f4f6;color:#6b7280;text-align:center;">%d</td>
			<td style="padding:10px 0;border-bottom:1px solid #f3f4f6;color:#111827;text-align:right;">%s</td>
		</tr>`, name, item.Quantity, formatARS(item.UnitPrice*float64(item.Quantity)))
	}

	var subtotal float64
	for _, item := range o.Items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}

	totalsRows := fmt.Sprintf(`
	<tr>
		<td colspan="2" style="padding:6px 0;color:#6b7280;font-size:13px;">Subtotal</td>
		<td style="padding:6px 0;color:#111827;text-align:right;font-size:13px;">%s</td>
	</tr>`, formatARS(subtotal))

	if o.ShippingCost > 0 {
		totalsRows += fmt.Sprintf(`
	<tr>
		<td colspan="2" style="padding:6px 0;color:#6b7280;font-size:13px;">Envío</td>
		<td style="padding:6px 0;color:#111827;text-align:right;font-size:13px;">%s</td>
	</tr>`, formatARS(o.ShippingCost))
	}

	if o.Discount > 0 {
		totalsRows += fmt.Sprintf(`
	<tr>
		<td colspan="2" style="padding:6px 0;color:#059669;font-size:13px;">Descuento transferencia</td>
		<td style="padding:6px 0;color:#059669;text-align:right;font-size:13px;">− %s</td>
	</tr>`, formatARS(o.Discount))
	}

	var entrega string
	if o.DeliveryMethod == "retirar" {
		entrega = "Retiro en local — Zona Pilar"
	} else {
		sd := o.ShippingDetails
		parts := []string{sd.Address, sd.City, sd.Province}
		var filtered []string
		for _, p := range parts {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		entrega = strings.Join(filtered, ", ")
		if sd.PostalCode != "" {
			entrega += " (" + sd.PostalCode + ")"
		}
	}

	pago := o.PaymentMethod
	switch pago {
	case "mercadopago":
		pago = "MercadoPago"
	case "transferencia":
		pago = "Transferencia bancaria"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width,initial-scale=1"/></head>
<body style="margin:0;padding:0;background:#f9fafb;font-family:system-ui,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f9fafb;padding:32px 16px;">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
  <tr>
    <td style="background:#0B1F3A;padding:28px 32px;">
      <p style="margin:0;color:#94a3b8;font-size:13px;">pooled.com.ar</p>
      <h1 style="margin:8px 0 4px;color:#ffffff;font-size:22px;">¡Pedido confirmado!</h1>
      <p style="margin:0;color:#64748b;font-size:13px;">Pedido <strong style="color:#94a3b8;">#%s</strong></p>
    </td>
  </tr>
  <tr><td style="padding:28px 32px;">
    <p style="margin:0 0 20px;color:#374151;font-size:15px;">Hola <strong>%s</strong>, recibimos tu pedido y lo estamos procesando.</p>
    <p style="margin:0 0 8px;font-weight:600;color:#0B1F3A;font-size:13px;text-transform:uppercase;letter-spacing:0.05em;">Productos</p>
    <table width="100%%" cellpadding="0" cellspacing="0" style="font-size:14px;">
      <tr>
        <th style="padding:8px 0;border-bottom:2px solid #e5e7eb;text-align:left;color:#6b7280;font-weight:500;">Producto</th>
        <th style="padding:8px 0;border-bottom:2px solid #e5e7eb;text-align:center;color:#6b7280;font-weight:500;">Cant.</th>
        <th style="padding:8px 0;border-bottom:2px solid #e5e7eb;text-align:right;color:#6b7280;font-weight:500;">Subtotal</th>
      </tr>
      %s
      %s
      <tr>
        <td colspan="2" style="padding:12px 0 4px;font-weight:700;color:#0B1F3A;font-size:15px;">Total</td>
        <td style="padding:12px 0 4px;font-weight:700;color:#0B1F3A;font-size:15px;text-align:right;">%s</td>
      </tr>
    </table>
    <table width="100%%" cellpadding="0" cellspacing="0" style="margin-top:24px;font-size:14px;">
      <tr>
        <td width="50%%" style="vertical-align:top;padding-right:12px;">
          <p style="margin:0 0 6px;font-weight:600;color:#0B1F3A;font-size:13px;text-transform:uppercase;letter-spacing:0.05em;">Entrega</p>
          <p style="margin:0;color:#374151;">%s</p>
        </td>
        <td width="50%%" style="vertical-align:top;">
          <p style="margin:0 0 6px;font-weight:600;color:#0B1F3A;font-size:13px;text-transform:uppercase;letter-spacing:0.05em;">Método de pago</p>
          <p style="margin:0;color:#374151;">%s</p>
        </td>
      </tr>
    </table>
    %s
    <p style="margin:28px 0 0;padding:16px;background:#f0fdf4;border-radius:8px;font-size:13px;color:#166534;">
      📦 Te avisamos cuando tu pedido esté en camino. Si tenés alguna consulta respondé este email o escribinos por WhatsApp +54 9 11 2342-7593.
    </p>
  </td></tr>
  <tr>
    <td style="padding:16px 32px;background:#f8fafc;border-top:1px solid #e5e7eb;text-align:center;">
      <p style="margin:0;font-size:12px;color:#9ca3af;">pooled.com.ar — Iluminación subacuática</p>
    </td>
  </tr>
</table>
</td></tr>
</table>
</body>
</html>`,
		orderRef, o.CustomerName,
		itemRows, totalsRows, formatARS(o.Total),
		entrega, pago, notasHTML(o),
	)
}

// SendShippingNotification avisa al cliente que su pedido fue despachado,
// incluyendo el número de seguimiento.
func (s *EmailService) SendShippingNotification(order *domain.Order) error {
	if !s.Enabled() {
		return nil
	}
	ref := strings.ToUpper(order.ID.Hex()[len(order.ID.Hex())-8:])
	subject := fmt.Sprintf("🚚 Tu pedido #%s está en camino — Pooled", ref)
	return s.Send(order.CustomerEmail, subject, buildShippingHTML(order))
}

// ── HTML: shipping notification ───────────────────────────────────────────────

func buildShippingHTML(o *domain.Order) string {
	orderRef := strings.ToUpper(o.ID.Hex()[len(o.ID.Hex())-8:])

	var destino string
	if o.DeliveryMethod == "retirar" {
		destino = "Retiro en local — Zona Pilar"
	} else {
		sd := o.ShippingDetails
		parts := []string{sd.Address, sd.City, sd.Province}
		var filtered []string
		for _, p := range parts {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		destino = strings.Join(filtered, ", ")
		if sd.PostalCode != "" {
			destino += " (" + sd.PostalCode + ")"
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width,initial-scale=1"/></head>
<body style="margin:0;padding:0;background:#f9fafb;font-family:system-ui,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f9fafb;padding:32px 16px;">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
  <tr>
    <td style="background:#0B1F3A;padding:28px 32px;">
      <p style="margin:0;color:#94a3b8;font-size:13px;">pooled.com.ar</p>
      <h1 style="margin:8px 0 4px;color:#ffffff;font-size:22px;">🚚 ¡Tu pedido está en camino!</h1>
      <p style="margin:0;color:#64748b;font-size:13px;">Pedido <strong style="color:#94a3b8;">#%s</strong></p>
    </td>
  </tr>
  <tr><td style="padding:28px 32px;">
    <p style="margin:0 0 20px;color:#374151;font-size:15px;">Hola <strong>%s</strong>, despachamos tu pedido.</p>
    <p style="margin:0 0 8px;font-weight:600;color:#0B1F3A;font-size:13px;text-transform:uppercase;letter-spacing:0.05em;">Número de seguimiento</p>
    <div style="margin:0 0 24px;">
      <span style="display:inline-block;background:#f3f4f6;border-radius:10px;padding:14px 28px;font-size:22px;font-weight:700;letter-spacing:2px;color:#0B1F3A;font-family:monospace;">%s</span>
    </div>
    <p style="margin:0 0 6px;font-weight:600;color:#0B1F3A;font-size:13px;text-transform:uppercase;letter-spacing:0.05em;">Destino</p>
    <p style="margin:0 0 24px;color:#374151;font-size:14px;">%s</p>
    <p style="margin:0;padding:16px;background:#eff6ff;border-radius:8px;font-size:13px;color:#1e40af;">
      Podés hacer el seguimiento de tu envío con ese número desde la aplicación de Andreani. Si tenés alguna consulta escribinos por WhatsApp +54 9 11 2342-7593.
    </p>
  </td></tr>
  <tr>
    <td style="padding:16px 32px;background:#f8fafc;border-top:1px solid #e5e7eb;text-align:center;">
      <p style="margin:0;font-size:12px;color:#9ca3af;">pooled.com.ar — Iluminación subacuática</p>
    </td>
  </tr>
</table>
</td></tr>
</table>
</body>
</html>`, orderRef, o.CustomerName, o.TrackingNumber, destino)
}

// SendOwnerNewOrder notifica a los dueños que llegó un pedido nuevo.
func (s *EmailService) SendOwnerNewOrder(order *domain.Order) {
	if !s.Enabled() || len(s.OwnerEmails) == 0 {
		return
	}
	ref := strings.ToUpper(order.ID.Hex()[len(order.ID.Hex())-8:])
	subject := fmt.Sprintf("🛒 Nuevo pedido #%s — %s", ref, order.CustomerName)
	html := buildOwnerNewOrderHTML(order)
	for _, email := range s.OwnerEmails {
		if err := s.Send(email, subject, html); err != nil {
			log.Printf("SendOwnerNewOrder: error enviando a %s: %v", email, err)
		}
	}
}

// SendOwnerStatusChange notifica a los dueños que cambió el estado de un pedido.
func (s *EmailService) SendOwnerStatusChange(order *domain.Order) {
	if !s.Enabled() || len(s.OwnerEmails) == 0 {
		return
	}
	ref := strings.ToUpper(order.ID.Hex()[len(order.ID.Hex())-8:])
	statusLabel := ownerStatusLabel(order.Status)
	subject := fmt.Sprintf("📋 Pedido #%s → %s — %s", ref, statusLabel, order.CustomerName)
	html := buildOwnerStatusChangeHTML(order)
	for _, email := range s.OwnerEmails {
		if err := s.Send(email, subject, html); err != nil {
			log.Printf("SendOwnerStatusChange: error enviando a %s: %v", email, err)
		}
	}
}

// SendOwnerNewUser notifica a los dueños que se registró un nuevo usuario.
func (s *EmailService) SendOwnerNewUser(user *domain.User) {
	if !s.Enabled() || len(s.OwnerEmails) == 0 {
		return
	}
	subject := fmt.Sprintf("👤 Nuevo registro — %s", user.Email)
	html := buildOwnerNewUserHTML(user)
	for _, email := range s.OwnerEmails {
		if err := s.Send(email, subject, html); err != nil {
			log.Printf("SendOwnerNewUser: error enviando a %s: %v", email, err)
		}
	}
}

func buildOwnerNewUserHTML(u *domain.User) string {
	registered := u.CreatedAt.Format("02/01/2006 15:04")
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="es">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f3f4f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f3f4f6;padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
        <tr><td style="background-color:#0ea5e9;padding:24px 32px;">
          <h1 style="margin:0;color:#ffffff;font-size:18px;font-weight:600;">👤 Nuevo usuario registrado</h1>
        </td></tr>
        <tr><td style="padding:28px 32px;">
          <p style="margin:0 0 16px;color:#374151;font-size:14px;">Se registró una nueva persona en la tienda.</p>
          <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="font-size:14px;color:#111827;">
            <tr>
              <td style="padding:8px 0;color:#6b7280;">Email</td>
              <td style="padding:8px 0;text-align:right;font-weight:600;">%s</td>
            </tr>
            <tr>
              <td style="padding:8px 0;color:#6b7280;border-top:1px solid #f3f4f6;">Fecha de registro</td>
              <td style="padding:8px 0;text-align:right;font-weight:600;border-top:1px solid #f3f4f6;">%s</td>
            </tr>
          </table>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, u.Email, registered)
}

func ownerStatusLabel(status string) string {
	labels := map[string]string{
		"pending":    "Pendiente",
		"paid":       "Pagado",
		"processing": "En proceso",
		"shipped":    "Enviado",
		"delivered":  "Entregado",
		"cancelled":  "Cancelado",
		"rejected":   "Rechazado",
	}
	if l, ok := labels[status]; ok {
		return l
	}
	return status
}

func ownerStatusColor(status string) string {
	colors := map[string]string{
		"pending":    "#f59e0b",
		"paid":       "#10b981",
		"processing": "#3b82f6",
		"shipped":    "#8b5cf6",
		"delivered":  "#059669",
		"cancelled":  "#ef4444",
		"rejected":   "#dc2626",
	}
	if c, ok := colors[status]; ok {
		return c
	}
	return "#6b7280"
}

// ── HTML: owner new order ─────────────────────────────────────────────────────

func buildOwnerNewOrderHTML(o *domain.Order) string {
	orderRef := strings.ToUpper(o.ID.Hex()[len(o.ID.Hex())-8:])

	var itemRows string
	for _, item := range o.Items {
		name := item.Name
		if name == "" {
			name = item.ProductID.Hex()
		}
		sku := ""
		if item.VariantSKU != "" {
			sku = fmt.Sprintf(" <span style='color:#6b7280;font-size:12px;'>(%s)</span>", item.VariantSKU)
		}
		itemRows += fmt.Sprintf(`
		<tr>
			<td style="padding:8px 0;border-bottom:1px solid #f3f4f6;color:#111827;font-size:14px;">%s%s</td>
			<td style="padding:8px 0;border-bottom:1px solid #f3f4f6;color:#6b7280;text-align:center;font-size:14px;">%d</td>
			<td style="padding:8px 0;border-bottom:1px solid #f3f4f6;color:#111827;text-align:right;font-size:14px;">%s</td>
		</tr>`, name, sku, item.Quantity, formatARS(item.UnitPrice*float64(item.Quantity)))
	}

	pago := o.PaymentMethod
	switch pago {
	case "mercadopago":
		pago = "MercadoPago"
	case "transferencia":
		pago = "Transferencia bancaria"
	}

	entrega := "Retiro en local"
	if o.DeliveryMethod != "retirar" {
		sd := o.ShippingDetails
		parts := []string{sd.Address, sd.City, sd.Province}
		var filtered []string
		for _, p := range parts {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		entrega = strings.Join(filtered, ", ")
		if sd.PostalCode != "" {
			entrega += " (" + sd.PostalCode + ")"
		}
	}

	dniRow := ""
	if o.DniCuit != "" {
		dniRow = fmt.Sprintf(`<tr><td style="color:#6b7280;font-size:13px;padding:3px 0;">DNI/CUIT</td><td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s</strong></td></tr>`, o.DniCuit)
	}
	facturaRow := ""
	if o.FacturaA != nil {
		facturaRow = fmt.Sprintf(`<tr><td style="color:#6b7280;font-size:13px;padding:3px 0;">Factura A</td><td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s — %s</strong></td></tr>`, o.FacturaA.RazonSocial, o.FacturaA.CUIT)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/></head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:system-ui,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f1f5f9;padding:32px 16px;">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:580px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.10);">
  <tr>
    <td style="background:#0B1F3A;padding:24px 32px;">
      <p style="margin:0;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:0.08em;">Notificación interna — Pooled</p>
      <h1 style="margin:8px 0 4px;color:#ffffff;font-size:22px;">🛒 Nuevo pedido recibido</h1>
      <p style="margin:0;color:#64748b;font-size:13px;">Pedido <strong style="color:#94a3b8;">#%s</strong></p>
    </td>
  </tr>
  <tr><td style="padding:28px 32px;">

    <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:24px;">
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;width:110px;">Cliente</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s</strong></td>
      </tr>
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Email</td>
        <td style="font-size:13px;padding:3px 0;"><a href="mailto:%s" style="color:#2563eb;">%s</a></td>
      </tr>
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Teléfono</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;">%s</td>
      </tr>
      %s
      %s
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Pago</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s</strong></td>
      </tr>
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Entrega</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;">%s</td>
      </tr>
    </table>

    <p style="margin:0 0 8px;font-weight:600;color:#0B1F3A;font-size:12px;text-transform:uppercase;letter-spacing:0.05em;">Productos</p>
    <table width="100%%" cellpadding="0" cellspacing="0">
      <tr>
        <th style="padding:6px 0;border-bottom:2px solid #e5e7eb;text-align:left;color:#6b7280;font-weight:500;font-size:12px;">Producto</th>
        <th style="padding:6px 0;border-bottom:2px solid #e5e7eb;text-align:center;color:#6b7280;font-weight:500;font-size:12px;">Cant.</th>
        <th style="padding:6px 0;border-bottom:2px solid #e5e7eb;text-align:right;color:#6b7280;font-weight:500;font-size:12px;">Subtotal</th>
      </tr>
      %s
      <tr>
        <td colspan="2" style="padding:10px 0 4px;font-weight:700;color:#0B1F3A;font-size:15px;">Total</td>
        <td style="padding:10px 0 4px;font-weight:700;color:#0B1F3A;font-size:15px;text-align:right;">%s</td>
      </tr>
    </table>

    %s

  </td></tr>
  <tr>
    <td style="padding:14px 32px;background:#f8fafc;border-top:1px solid #e5e7eb;text-align:center;">
      <p style="margin:0;font-size:11px;color:#9ca3af;">Esta notificación es solo para uso interno — pooled.com.ar</p>
    </td>
  </tr>
</table>
</td></tr>
</table>
</body>
</html>`,
		orderRef,
		o.CustomerName,
		o.CustomerEmail, o.CustomerEmail,
		o.CustomerPhone,
		dniRow, facturaRow,
		pago, entrega,
		itemRows,
		formatARS(o.Total),
		notasHTML(o),
	)
}

// ── HTML: owner status change ─────────────────────────────────────────────────

func buildOwnerStatusChangeHTML(o *domain.Order) string {
	orderRef := strings.ToUpper(o.ID.Hex()[len(o.ID.Hex())-8:])
	statusLabel := ownerStatusLabel(o.Status)
	statusColor := ownerStatusColor(o.Status)

	trackingRow := ""
	if o.TrackingNumber != "" {
		trackingRow = fmt.Sprintf(`
		<tr>
			<td style="color:#6b7280;font-size:13px;padding:3px 0;width:130px;">N° de seguimiento</td>
			<td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s</strong></td>
		</tr>`, o.TrackingNumber)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/></head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:system-ui,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f1f5f9;padding:32px 16px;">
<tr><td align="center">
<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.10);">
  <tr>
    <td style="background:#0B1F3A;padding:24px 32px;">
      <p style="margin:0;color:#94a3b8;font-size:12px;text-transform:uppercase;letter-spacing:0.08em;">Notificación interna — Pooled</p>
      <h1 style="margin:8px 0 4px;color:#ffffff;font-size:22px;">📋 Estado actualizado</h1>
      <p style="margin:0;color:#64748b;font-size:13px;">Pedido <strong style="color:#94a3b8;">#%s</strong></p>
    </td>
  </tr>
  <tr><td style="padding:28px 32px;">

    <div style="text-align:center;margin-bottom:28px;">
      <span style="display:inline-block;background:%s;color:#ffffff;border-radius:20px;padding:8px 28px;font-size:16px;font-weight:700;letter-spacing:0.03em;">%s</span>
    </div>

    <table width="100%%" cellpadding="0" cellspacing="0">
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;width:130px;">Cliente</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;"><strong>%s</strong></td>
      </tr>
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Email</td>
        <td style="font-size:13px;padding:3px 0;"><a href="mailto:%s" style="color:#2563eb;">%s</a></td>
      </tr>
      <tr>
        <td style="color:#6b7280;font-size:13px;padding:3px 0;">Total del pedido</td>
        <td style="color:#111827;font-size:13px;padding:3px 0;font-weight:600;">%s</td>
      </tr>
      %s
    </table>

  </td></tr>
  <tr>
    <td style="padding:14px 32px;background:#f8fafc;border-top:1px solid #e5e7eb;text-align:center;">
      <p style="margin:0;font-size:11px;color:#9ca3af;">Esta notificación es solo para uso interno — pooled.com.ar</p>
    </td>
  </tr>
</table>
</td></tr>
</table>
</body>
</html>`,
		orderRef,
		statusColor, statusLabel,
		o.CustomerName,
		o.CustomerEmail, o.CustomerEmail,
		formatARS(o.Total),
		trackingRow,
	)
}

func notasHTML(o *domain.Order) string {
	if o.Notes == "" {
		return ""
	}
	return fmt.Sprintf(`<p style="margin:20px 0 0;padding:12px 16px;background:#fefce8;border-left:3px solid #eab308;border-radius:4px;font-size:13px;color:#713f12;"><strong>Notas:</strong> %s</p>`, o.Notes)
}

func formatARS(amount float64) string {
	intPart := int64(amount)
	decPart := int(((amount - float64(intPart)) * 100) + 0.5)
	s := fmt.Sprintf("%d", intPart)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return fmt.Sprintf("$ %s,%02d", result, decPart)
}
