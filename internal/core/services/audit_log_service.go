package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLogService struct {
	repo *repositories.AuditLogRepositoryMongo
}

func NewAuditLogService(repo *repositories.AuditLogRepositoryMongo) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// routeLabels mapea "MÉTODO patrón-de-ruta" (tal cual lo da gin's c.FullPath())
// a una descripción legible. Las rutas sin entrada acá caen a un label genérico
// armado desde el método + ruta.
var routeLabels = map[string]string{
	"POST /admin/products":                          "Creó un producto",
	"PUT /admin/products/:id":                       "Editó un producto",
	"DELETE /admin/products/:id":                    "Eliminó un producto",
	"PATCH /admin/products/bulk-price":              "Actualizó precios en lote",
	"PATCH /admin/products/:id/visibility":          "Cambió visibilidad de un producto",
	"PATCH /admin/products/:id/sort-order":          "Reordenó un producto",
	"PATCH /admin/products/:id/variants/:sku/stock": "Ajustó stock de una variante",

	"POST /admin/kits":                          "Creó un kit",
	"PUT /admin/kits/:id":                       "Editó un kit",
	"DELETE /admin/kits/:id":                    "Eliminó un kit",
	"PATCH /admin/kits/:id/visibility":          "Cambió visibilidad de un kit",
	"PATCH /admin/kits/:id/sort-order":          "Reordenó un kit",
	"PATCH /admin/kits/bulk-price":              "Actualizó precios de kits en lote",
	"POST /admin/kits/refresh-suggested-prices": "Actualizó kits al precio sugerido de sus productos",

	"POST /admin/coupons":       "Creó un cupón",
	"PUT /admin/coupons/:id":    "Editó un cupón",
	"DELETE /admin/coupons/:id": "Eliminó un cupón",

	"PATCH /admin/orders/:id/status":        "Cambió el estado de un pedido",
	"DELETE /admin/orders/:id":              "Eliminó un pedido",
	"POST /admin/orders/:id/invoice":        "Subió una factura",
	"POST /admin/orders/:id/invoice/resend": "Reenvió una factura",

	"DELETE /admin/users/:id":          "Eliminó un usuario",
	"POST /admin/users/invite-admin":   "Invitó a un nuevo administrador",
	"PATCH /admin/users/:id/role":      "Cambió el rol de un usuario",
	"POST /admin/bootstrap-superadmin": "Se auto-promovió a superadmin (bootstrap inicial)",

	"PUT /admin/settings/abandoned-cart-email": "Editó la plantilla del mail de carrito abandonado",
	"POST /admin/email-templates":              "Creó una plantilla de mail",
	"DELETE /admin/email-templates/:id":        "Eliminó una plantilla de mail",

	"POST /admin/cart-links": "Generó un link de carrito",
}

func labelFor(method, route string) string {
	// c.FullPath() incluye el prefijo del grupo de rutas (ej. "/api/v1"); la
	// tabla de labels está escrita sin él, así que lo sacamos antes de buscar.
	trimmed := strings.TrimPrefix(route, "/api/v1")
	if l, ok := routeLabels[method+" "+trimmed]; ok {
		return l
	}
	return method + " " + trimmed
}

// fieldLabelsEs traduce nombres de campos del body a español, para el detalle.
var fieldLabelsEs = map[string]string{
	"status": "Estado", "name": "Nombre", "email": "Email",
	"visible": "Visible", "active": "Activo",
	"tracking_number": "Seguimiento", "discount_percent": "Descuento",
	"base_price": "Precio", "stock": "Stock", "code": "Código",
}

// statusLabelsEs traduce estados de pedido a español, para el detalle.
var statusLabelsEs = map[string]string{
	"pending": "Pendiente", "paid": "Pagado", "processing": "Procesando",
	"shipped": "Enviado", "delivered": "Entregado",
	"cancelled": "Cancelado", "rejected": "Rechazado",
}

// humanizeDetails convierte el JSON crudo del body ({"status":"cancelled"})
// en texto legible ("Estado: Cancelado"), para que lo entienda alguien que
// no lee código. Si no se puede parsear, se devuelve tal cual vino.
func humanizeDetails(raw string) string {
	if raw == "" {
		return ""
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return raw
	}
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		label, ok := fieldLabelsEs[k]
		if !ok {
			label = k
		}
		var valStr string
		if s, ok := parsed[k].(string); ok {
			if k == "status" {
				if es, found := statusLabelsEs[s]; found {
					s = es
				}
			}
			valStr = s
		} else if b, ok := parsed[k].(bool); ok {
			if b {
				valStr = "Sí"
			} else {
				valStr = "No"
			}
		} else {
			b, _ := json.Marshal(parsed[k])
			valStr = string(b)
		}
		if valStr == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", label, valStr))
	}
	return strings.Join(parts, " · ")
}

// LogAction registra una acción de escritura de un admin. Pensado para
// llamarse en background (goroutine) desde el middleware — nunca debe
// bloquear ni romper la request que la originó.
func (s *AuditLogService) LogAction(ctx context.Context, adminID primitive.ObjectID, adminEmail, method, route, path, details string, statusCode int) error {
	entry := &domain.AdminAuditLog{
		AdminID:    adminID,
		AdminEmail: adminEmail,
		Method:     method,
		Route:      route,
		Path:       path,
		Label:      labelFor(method, route),
		Details:    humanizeDetails(details),
		StatusCode: statusCode,
	}
	return s.repo.Create(ctx, entry)
}

func (s *AuditLogService) List(ctx context.Context, adminID primitive.ObjectID, from, to time.Time, page, limit int64) ([]*domain.AdminAuditLog, int64, error) {
	skip := (page - 1) * limit
	logs, err := s.repo.List(ctx, adminID, from, to, skip, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, adminID, from, to)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

type AdminOption struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (s *AuditLogService) ListAdmins(ctx context.Context) ([]AdminOption, error) {
	rows, err := s.repo.DistinctAdmins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AdminOption, 0, len(rows))
	for _, r := range rows {
		out = append(out, AdminOption{ID: r.AdminID.Hex(), Email: r.AdminEmail})
	}
	return out, nil
}
