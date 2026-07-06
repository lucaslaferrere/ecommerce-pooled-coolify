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
	Attr3           string  `bson:"attr3,omitempty" json:"attr3,omitempty"`
	Stock           int     `bson:"stock" json:"stock"`
	PriceAdjustment float64 `bson:"price_adjustment" json:"price_adjustment"` // Ajuste al precio base
}

// Spec representa un par clave/valor de la ficha técnica del producto
type Spec struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

// MainSpec agrega un campo "meaning" (explicación B2C) a la ficha técnica
type MainSpec struct {
	Key     string `bson:"key" json:"key"`
	Value   string `bson:"value" json:"value"`
	Meaning string `bson:"meaning" json:"meaning"`
}

// Benefit representa un beneficio del producto (contenido B2C)
type Benefit struct {
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
}

// Product representa la entidad de producto en la aplicación
type Product struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name" json:"name"`
	Subtitle        string             `bson:"subtitle,omitempty" json:"subtitle,omitempty"`
	Description     string             `bson:"description" json:"description"`
	BasePrice       float64            `bson:"base_price" json:"base_price"`
	DiscountPercent float64            `bson:"discount_percent" json:"discount_percent"`
	Stock           int                `bson:"stock" json:"stock"`
	Category        string             `bson:"category" json:"category"`
	Brand           string             `bson:"brand" json:"brand"`
	Images          []string           `bson:"images" json:"images"`
	Variants        []Variant          `bson:"variants" json:"variants"`
	VariantLabels   []string           `bson:"variant_labels,omitempty" json:"variant_labels,omitempty"`
	Specs           []Spec             `bson:"specs" json:"specs"`
	MainSpecs       []MainSpec         `bson:"main_specs,omitempty" json:"main_specs,omitempty"`
	Benefits        []Benefit          `bson:"benefits,omitempty" json:"benefits,omitempty"`
	SortOrder       int                `bson:"sort_order" json:"sort_order"`
	Visible         *bool              `bson:"visible,omitempty" json:"visible,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
