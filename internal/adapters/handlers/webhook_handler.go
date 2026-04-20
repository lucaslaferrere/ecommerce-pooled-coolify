package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookHandler procesa notificaciones entrantes de Mercado Pago.
type WebhookHandler struct {
	paymentService *services.PaymentService
	orderService   *services.OrderService
	webhookSecret  string
}

// NewWebhookHandler crea una nueva instancia de WebhookHandler.
func NewWebhookHandler(
	paymentService *services.PaymentService,
	orderService *services.OrderService,
	webhookSecret string,
) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
		orderService:   orderService,
		webhookSecret:  webhookSecret,
	}
}

// mpWebhookBody es el cuerpo del webhook v2 de Mercado Pago.
type mpWebhookBody struct {
	Action   string `json:"action"`
	Type     string `json:"type"`
	LiveMode bool   `json:"live_mode"`
	Data     struct {
		ID string `json:"id"`
	} `json:"data"`
}

// HandleMercadoPago procesa POST /api/v1/webhooks/mercadopago.
//
// Flujo:
//  1. Valida la firma HMAC-SHA256 (si MP_WEBHOOK_SECRET está configurado).
//  2. Extrae el ID de pago del cuerpo (webhook v2) o de los query params (IPN legacy).
//  3. Consulta la API de MP para verificar el estado real del pago.
//  4. Si el pago está "approved", actualiza la orden a estado "paid".
//
// Siempre responde 200 OK a MP para evitar reintentos innecesarios, incluso
// cuando el evento es ignorado o ya fue procesado.
func (h *WebhookHandler) HandleMercadoPago(c *gin.Context) {
	// 1. Validación de firma (solo si el secret está configurado)
	if h.webhookSecret != "" {
		if err := h.validateSignature(c); err != nil {
			log.Printf("webhook mp: firma inválida: %v", err)
			// Responder 200 para que MP no reintente; logueamos el incidente.
			c.JSON(http.StatusOK, gin.H{"message": "firma inválida, evento ignorado"})
			return
		}
	}

	// 2. Parsear el cuerpo como webhook v2
	var body mpWebhookBody
	_ = c.ShouldBindJSON(&body) // ignorar error; completamos con query params si falla

	// Determinar tipo y ID de pago — cuerpo tiene prioridad, query params como fallback (IPN)
	eventType := body.Type
	paymentID := body.Data.ID

	if paymentID == "" {
		paymentID = c.Query("id")
	}
	if eventType == "" {
		eventType = c.Query("topic") // IPN usa "topic" en lugar de "type"
	}

	// Solo procesar eventos de tipo "payment"
	if eventType != "payment" || paymentID == "" {
		c.JSON(http.StatusOK, gin.H{"message": "evento ignorado"})
		return
	}

	log.Printf("webhook mp: procesando pago id=%s live=%v", paymentID, body.LiveMode)

	// 3. Consultar estado real del pago en la API de MP
	paymentInfo, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		log.Printf("webhook mp: error consultando pago %s: %v", paymentID, err)
		// 200 para que MP no reintente en bucle; el error se investiga en logs
		c.JSON(http.StatusOK, gin.H{"message": "error consultando pago, reintentando después"})
		return
	}

	log.Printf("webhook mp: pago %s → status=%s detail=%s ref=%s",
		paymentID, paymentInfo.Status, paymentInfo.StatusDetail, paymentInfo.ExternalReference)

	// 4. Solo nos interesa el estado "approved"
	if paymentInfo.Status != "approved" {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("pago %s, ignorado", paymentInfo.Status)})
		return
	}

	// 5. Localizar la orden por external_reference (= Order.ID.Hex())
	orderID, err := primitive.ObjectIDFromHex(paymentInfo.ExternalReference)
	if err != nil {
		log.Printf("webhook mp: external_reference inválida %q: %v", paymentInfo.ExternalReference, err)
		c.JSON(http.StatusOK, gin.H{"message": "referencia de orden inválida"})
		return
	}

	// 6. Confirmar pago (idempotente: si ya está "paid" no hace nada)
	if err := h.orderService.ConfirmPayment(c.Request.Context(), orderID, paymentInfo.ID); err != nil {
		log.Printf("webhook mp: error confirmando orden %s: %v", orderID.Hex(), err)
		c.JSON(http.StatusOK, gin.H{"message": "error actualizando orden"})
		return
	}

	log.Printf("webhook mp: orden %s marcada como pagada (pago %s)", orderID.Hex(), paymentInfo.ID)
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// validateSignature verifica la firma HMAC-SHA256 de Mercado Pago.
//
// MP envía en el header x-signature: ts=<timestamp>,v1=<hash>
// El manifest que se firma es: id:<data.id>;request-id:<x-request-id>;ts:<ts>
// donde data.id proviene del query param (siempre presente en webhooks v2).
func (h *WebhookHandler) validateSignature(c *gin.Context) error {
	xSig := c.GetHeader("x-signature")
	xReqID := c.GetHeader("x-request-id")
	dataID := c.Query("data.id")

	if xSig == "" {
		return fmt.Errorf("header x-signature ausente")
	}

	var ts, v1 string
	for _, part := range strings.Split(xSig, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}

	if ts == "" || v1 == "" {
		return fmt.Errorf("x-signature con formato inválido: %q", xSig)
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s", dataID, xReqID, ts)

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(v1)) {
		return fmt.Errorf("firma no coincide (manifest=%q)", manifest)
	}
	return nil
}
