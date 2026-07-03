package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invoice es la factura (PDF) asociada a un pedido. El PDF se guarda como base64
// en Content para no depender del filesystem (que se pierde en redeploys).
type Invoice struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id,omitempty"`
	OrderID   primitive.ObjectID `bson:"order_id"        json:"order_id"`
	Filename  string             `bson:"filename"        json:"filename"`
	Content   string             `bson:"content"         json:"-"` // base64; nunca se expone por JSON
	SentAt    *time.Time         `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"      json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"      json:"updated_at"`
}
