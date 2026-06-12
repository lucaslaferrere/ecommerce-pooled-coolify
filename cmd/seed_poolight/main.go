package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	BasePrice   float64            `bson:"base_price"`
	Stock       int                `bson:"stock"`
	Category    string             `bson:"category"`
	Brand       string             `bson:"brand"`
	Images      []string           `bson:"images"`
	Variants    []interface{}      `bson:"variants"`
	Specs       []interface{}      `bson:"specs"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func main() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb://admin:password@127.0.0.1:27017/ecommerce?authSource=admin",
	))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	col := client.Database("ecommerce").Collection("products")
	now := time.Now()

	products := []Product{
		{
			Name: "POOLIGHT Wall Mounted Plastic RGBW", BasePrice: 100575.40,
			Description: "Luminaria de pared para piscina, cuerpo plástico, luz RGBW.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Wall Mounted Plastic Blanco", BasePrice: 74943.32,
			Description: "Luminaria de pared para piscina, cuerpo plástico, luz blanca.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Wall Mounted SS316 RGBW", BasePrice: 125800.62,
			Description: "Luminaria de pared para piscina, acero inoxidable SS316, luz RGBW.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Wall Mounted SS316 Blanco", BasePrice: 92845.09,
			Description: "Luminaria de pared para piscina, acero inoxidable SS316, luz blanca.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Mini 1.5\" Plastic RGBW", BasePrice: 131659.38,
			Description: "Luminaria mini 1.5\" para piscina, cuerpo plástico, luz RGBW.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Mini 1.5\" Plastic Blanco", BasePrice: 103179.29,
			Description: "Luminaria mini 1.5\" para piscina, cuerpo plástico, luz blanca.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Mini 1.5\" SS316 RGBW", BasePrice: 157494.89,
			Description: "Luminaria mini 1.5\" para piscina, acero inoxidable SS316, luz RGBW.", Category: "luminarias",
		},
		{
			Name: "POOLIGHT Mini 1.5\" SS316 Blanco", BasePrice: 129014.80,
			Description: "Luminaria mini 1.5\" para piscina, acero inoxidable SS316, luz blanca.", Category: "luminarias",
		},
	}

	for i := range products {
		products[i].ID = primitive.NewObjectID()
		products[i].Brand = "Pooled"
		products[i].Stock = 10
		products[i].Images = []string{}
		products[i].Variants = []interface{}{}
		products[i].Specs = []interface{}{}
		products[i].CreatedAt = now
		products[i].UpdatedAt = now
	}

	docs := make([]interface{}, len(products))
	for i, p := range products {
		docs[i] = p
	}

	res, err := col.InsertMany(ctx, docs)
	if err != nil {
		log.Fatal("Error insertando:", err)
	}
	log.Printf("✓ %d productos POOLIGHT insertados correctamente", len(res.InsertedIDs))
}
