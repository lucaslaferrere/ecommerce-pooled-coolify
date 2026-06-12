package domain

import "time"

type CartLinkItem struct {
	ProductID  string `bson:"product_id" json:"product_id"`
	VariantSKU string `bson:"variant_sku" json:"variant_sku"`
	Name       string `bson:"name" json:"name"`
	ImageURL   string `bson:"image_url" json:"image_url"`
	UnitPrice  float64 `bson:"unit_price" json:"unit_price"`
	Quantity   int    `bson:"quantity" json:"quantity"`
	Type       string `bson:"type" json:"type"` // "product" | "kit"
}

type CartLink struct {
	Token     string         `bson:"token" json:"token"`
	Items     []CartLinkItem `bson:"items" json:"items"`
	CreatedAt time.Time      `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time      `bson:"expires_at" json:"expires_at"`
}
