package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Variant representa una variante de producto (ej: talla, color, etc)
type Variant struct {
	SKU             string  `bson:"sku" json:"sku"`
	Color           string  `bson:"color" json:"color"`
	Size            string  `bson:"size" json:"size"`
	Stock           int     `bson:"stock" json:"stock"`
	PriceAdjustment float64 `bson:"price_adjustment" json:"price_adjustment"` // Ajuste al precio base
}

// Product representa la entidad de producto en la aplicación
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	BasePrice   float64            `bson:"base_price" json:"base_price"`
	Category    string             `bson:"category" json:"category"`
	Brand       string             `bson:"brand" json:"brand"`
	Images      []string           `bson:"images" json:"images"`     // Array de URLs de imágenes
	Variants    []Variant          `bson:"variants" json:"variants"` // Slice de variantes
	CreatedAt   time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
