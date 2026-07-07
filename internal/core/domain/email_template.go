package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmailTemplate is a reusable body for the abandoned-cart recovery email.
type EmailTemplate struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Body      string             `bson:"body" json:"body"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
