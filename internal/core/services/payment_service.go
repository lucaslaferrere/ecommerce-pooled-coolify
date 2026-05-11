package services

import (
	"context"
	"fmt"
	"strconv"

	"ecommerce-pooled/internal/core/domain"

	mpConfig "github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// PaymentPreference contiene los datos de la preferencia creada en Mercado Pago.
type PaymentPreference struct {
	PreferenceID     string
	InitPoint        string
	SandboxInitPoint string
}

// PaymentInfo contiene el estado de un pago consultado a la API de Mercado Pago.
type PaymentInfo struct {
	ID                string
	Status            string // approved | rejected | pending | in_process | cancelled ...
	StatusDetail      string
	ExternalReference string // corresponde al Order.ID.Hex()
}

// PaymentService gestiona la integración con la API de Mercado Pago.
// Cuando MPAccessToken está vacío, el servicio opera en modo desactivado:
// CreatePreference devuelve una preferencia vacía sin error, lo que permite
// continuar el flujo de desarrollo/test sin credenciales reales.
type PaymentService struct {
	mpCfg      *mpConfig.Config
	successURL string
	failureURL string
	pendingURL string
	webhookURL string
	currencyID string
	enabled    bool
}

// NewPaymentService construye el servicio. Si accessToken está vacío, retorna
// un servicio desactivado (sin error) para que el servidor pueda arrancar sin
// credenciales de MP.
func NewPaymentService(
	accessToken, webhookSecret string,
	successURL, failureURL, pendingURL string,
	appURL, currencyID string,
) (*PaymentService, error) {
	if accessToken == "" {
		return &PaymentService{enabled: false}, nil
	}

	cfg, err := mpConfig.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("error inicializando Mercado Pago: %w", err)
	}

	return &PaymentService{
		mpCfg:      cfg,
		successURL: successURL,
		failureURL: failureURL,
		pendingURL: pendingURL,
		webhookURL: appURL + "/api/v1/webhooks/mercadopago",
		currencyID: currencyID,
		enabled:    true,
	}, nil
}

// IsEnabled informa si el servicio está activo (tiene credenciales de MP).
func (s *PaymentService) IsEnabled() bool {
	return s.enabled
}

// CreatePreference genera una Preferencia de Mercado Pago para la orden dada.
// Retorna una preferencia vacía sin error cuando el servicio está desactivado,
// de modo que el checkout sigue funcionando en entornos sin credenciales.
func (s *PaymentService) CreatePreference(ctx context.Context, order *domain.Order) (*PaymentPreference, error) {
	if !s.enabled {
		return &PaymentPreference{}, nil
	}

	items := make([]preference.ItemRequest, len(order.Items))
	for i, item := range order.Items {
		itemID := item.VariantSKU
		if itemID == "" {
			itemID = item.ProductID.Hex()
		}
		title := item.VariantSKU
		if title == "" {
			title = "Producto " + item.ProductID.Hex()
		}
		unitPrice := item.UnitPrice
		if unitPrice <= 0 {
			unitPrice = 1
		}
		items[i] = preference.ItemRequest{
			ID:         itemID,
			Title:      title,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			CurrencyID: s.currencyID,
		}
	}

	req := preference.Request{
		Items:             items,
		ExternalReference: order.ID.Hex(),
		NotificationURL:   s.webhookURL,
		BackURLs: &preference.BackURLsRequest{
			Success: s.successURL,
			Failure: s.failureURL,
			Pending: s.pendingURL,
		},
	}

	client := preference.NewClient(s.mpCfg)
	result, err := client.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creando preferencia de pago: %w", err)
	}

	return &PaymentPreference{
		PreferenceID:     result.ID,
		InitPoint:        result.InitPoint,
		SandboxInitPoint: result.SandboxInitPoint,
	}, nil
}

// GetPayment consulta el estado real de un pago a la API de Mercado Pago.
// El paymentIDStr es el ID numérico del pago como string (tal como llega en el webhook).
func (s *PaymentService) GetPayment(ctx context.Context, paymentIDStr string) (*PaymentInfo, error) {
	if !s.enabled {
		return nil, ErrPaymentNotConfigured
	}

	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		return nil, fmt.Errorf("payment ID inválido %q: %w", paymentIDStr, err)
	}

	client := payment.NewClient(s.mpCfg)
	result, err := client.Get(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("error consultando pago %d en MP: %w", paymentID, err)
	}

	return &PaymentInfo{
		ID:                strconv.Itoa(result.ID),
		Status:            result.Status,
		StatusDetail:      result.StatusDetail,
		ExternalReference: result.ExternalReference,
	}, nil
}
