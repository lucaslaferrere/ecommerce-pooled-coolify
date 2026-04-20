package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CartItem representa un elemento en el carrito de compras
type CartItem struct {
	ProductID  primitive.ObjectID `bson:"product_id" json:"product_id"`
	VariantSKU string             `bson:"variant_sku" json:"variant_sku"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	UnitPrice  float64            `bson:"unit_price,omitempty" json:"unit_price,omitempty"`
}
