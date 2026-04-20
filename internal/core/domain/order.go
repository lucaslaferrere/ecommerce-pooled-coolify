package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShippingDetails contiene la información de entrega de un pedido
type ShippingDetails struct {
	Address    string `bson:"address" json:"address"`
	City       string `bson:"city" json:"city"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
}

// Order representa la entidad de orden/pedido en la aplicación
type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	Items           []CartItem         `bson:"items" json:"items"`
	Total           float64            `bson:"total" json:"total"`
	Status          string             `bson:"status" json:"status"` // pending | paid | processing | shipped | delivered | cancelled
	ShippingDetails ShippingDetails    `bson:"shipping_details" json:"shipping_details"`
	PreferenceID    string             `bson:"preference_id,omitempty" json:"preference_id,omitempty"` // ID de preferencia Mercado Pago
	PaymentID       string             `bson:"payment_id,omitempty" json:"payment_id,omitempty"`       // ID de pago confirmado por MP
	CreatedAt       time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
