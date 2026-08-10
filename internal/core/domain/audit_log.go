package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminAuditLog registra una acción de escritura (crear/editar/borrar) hecha
// por un admin, para poder ver después quién hizo qué y cuándo.
type AdminAuditLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AdminID    primitive.ObjectID `bson:"admin_id" json:"admin_id"`
	AdminEmail string             `bson:"admin_email" json:"admin_email"`
	Method     string             `bson:"method" json:"method"`
	Route      string             `bson:"route" json:"route"` // patrón de ruta, ej. /admin/orders/:id/status
	Path       string             `bson:"path" json:"path"`   // ruta resuelta, ej. /admin/orders/64f.../status
	Label      string             `bson:"label" json:"label"` // descripción legible, ej. "Cambió el estado de un pedido"
	Details    string             `bson:"details,omitempty" json:"details,omitempty"`
	StatusCode int                `bson:"status_code" json:"status_code"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
