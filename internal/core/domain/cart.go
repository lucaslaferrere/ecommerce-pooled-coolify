package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Cart es el carrito persistido de un usuario logueado. Se sincroniza desde el
// frontend en cada cambio y se vacía cuando el usuario concreta una compra.
type Cart struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"                json:"id,omitempty"`
	UserID             primitive.ObjectID `bson:"user_id"                      json:"user_id"`
	Items              []CartItem         `bson:"items"                        json:"items"`
	UpdatedAt          time.Time          `bson:"updated_at"                   json:"updated_at"`
	LastReminderSentAt *time.Time         `bson:"last_reminder_sent_at,omitempty" json:"last_reminder_sent_at,omitempty"`
}
