package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WizardRecommendation struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	PoolSize         string             `bson:"pool_size" json:"pool_size"`
	PoolType         string             `bson:"pool_type" json:"pool_type"`
	UsageType        string             `bson:"usage_type" json:"usage_type"`
	ControlType      string             `bson:"control_type" json:"control_type"`
	RecommendedKitID string             `bson:"recommended_kit_id" json:"recommended_kit_id"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at,omitempty"`
}
