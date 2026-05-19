package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Kit struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name"`
	Slug          string             `bson:"slug" json:"slug"`
	Description   string             `bson:"description" json:"description"`
	Price         float64            `bson:"price" json:"price"`
	OriginalPrice *float64           `bson:"original_price,omitempty" json:"original_price,omitempty"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
	PoolSize      string             `bson:"pool_size" json:"pool_size"`
	Line          string             `bson:"line,omitempty" json:"line,omitempty"`
	Materials     []string           `bson:"materials,omitempty" json:"materials,omitempty"`
	Uso           string             `bson:"uso,omitempty" json:"uso,omitempty"`
	ProductIDs    []string           `bson:"product_ids" json:"product_ids"`
	Featured      bool               `bson:"featured" json:"featured"`
	SortOrder     int                `bson:"sort_order" json:"sort_order"`
	Visible       *bool              `bson:"visible,omitempty" json:"visible,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
