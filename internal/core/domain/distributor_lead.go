package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DistributorLead struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FullName  string             `bson:"full_name" json:"full_name"`
	Company   string             `bson:"company" json:"company"`
	City      string             `bson:"city" json:"city"`
	Phone     string             `bson:"phone" json:"phone"`
	Email     string             `bson:"email" json:"email"`
	Message   string             `bson:"message" json:"message"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at,omitempty"`
}
