package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Coupon represents a discount coupon.
type Coupon struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"      json:"id,omitempty"`
	Code            string             `bson:"code"               json:"code"`
	DiscountPercent float64            `bson:"discount_percent"   json:"discount_percent"`
	Active          bool               `bson:"active"             json:"active"`
	ExpiresAt       *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"         json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"         json:"updated_at"`
}
