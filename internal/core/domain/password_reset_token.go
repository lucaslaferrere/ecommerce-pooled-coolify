package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PasswordResetToken almacena un código temporal de recuperación de contraseña
type PasswordResetToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"  json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"        json:"user_id"`
	Email     string             `bson:"email"          json:"email"`
	Code      string             `bson:"code"           json:"-"`
	ExpiresAt time.Time          `bson:"expires_at"     json:"expires_at"`
	Used      bool               `bson:"used"           json:"used"`
	CreatedAt time.Time          `bson:"created_at"     json:"created_at"`
}
