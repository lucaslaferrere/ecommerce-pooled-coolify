package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Coupon represents a discount coupon.
// MaxUses limits how many paid orders may redeem the coupon in total (0 = unlimited).
// MaxUsesPerUser limits redemptions per customer account (0 = unlimited).
type Coupon struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"      json:"id,omitempty"`
	Code            string             `bson:"code"               json:"code"`
	DiscountPercent float64            `bson:"discount_percent"   json:"discount_percent"`
	Active          bool               `bson:"active"             json:"active"`
	MaxUses         int                `bson:"max_uses"           json:"max_uses"`
	MaxUsesPerUser  int                `bson:"max_uses_per_user"  json:"max_uses_per_user"`
	ExpiresAt       *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"         json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"         json:"updated_at"`

	// UsedCount es un campo calculado (no se persiste): cuántas órdenes pagadas
	// usaron este cupón. Se completa al listar para mostrarlo en el admin.
	UsedCount int `bson:"-" json:"used_count"`
}
