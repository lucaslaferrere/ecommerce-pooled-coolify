package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa la entidad de usuario en la aplicación
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"password_hash,omitempty"`
	Role         string             `bson:"role" json:"role"` // "admin" or "client"
	CreatedAt    time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
