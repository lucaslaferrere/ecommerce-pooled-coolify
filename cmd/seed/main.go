package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ecommerce-pooled/internal/adapters/database"
	"ecommerce-pooled/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Variant struct {
	SKU             string  `bson:"sku"`
	Color           string  `bson:"color"`
	Size            string  `bson:"size"`
	Stock           int     `bson:"stock"`
	PriceAdjustment float64 `bson:"price_adjustment"`
}

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	BasePrice   float64            `bson:"base_price"`
	Category    string             `bson:"category"`
	Brand       string             `bson:"brand"`
	Images      []string           `bson:"images"`
	Variants    []Variant          `bson:"variants"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type Kit struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name"`
	Slug          string             `bson:"slug"`
	Description   string             `bson:"description"`
	Price         float64            `bson:"price"`
	OriginalPrice *float64           `bson:"original_price,omitempty"`
	ImageURL      string             `bson:"image_url"`
	PoolSize      string             `bson:"pool_size"`
	Featured      bool               `bson:"featured"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

func ptr(f float64) *float64 { return &f }

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := database.NewClient(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	defer client.Close(context.Background())

	db := client.GetDatabase()
	now := time.Now()

	// ── Admin user ────────────────────────────────────────────────────────────
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error generando bcrypt hash: %v", err)
	}

	users := db.Collection("users")
	filter := bson.M{"email": "admin@pooled.com.ar"}
	update := bson.M{
		"$set": bson.M{
			"password_hash": string(hash),
			"role":          "admin",
			"updated_at":    now,
		},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"email":      "admin@pooled.com.ar",
			"created_at": now,
		},
	}
	result, err := users.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Fatalf("Error al hacer upsert del admin: %v", err)
	}
	if result.UpsertedCount > 0 {
		fmt.Println(">>> Admin creado: admin@pooled.com.ar / password123")
	} else {
		fmt.Println(">>> Admin actualizado: admin@pooled.com.ar / password123")
	}

	// ── Productos ─────────────────────────────────────────────────────────────
	db.Collection("products").DeleteMany(ctx, bson.M{})
	db.Collection("kits").DeleteMany(ctx, bson.M{})

	products := []interface{}{
		Product{
			Name: "Luminaria LED RGB 12W", Description: "Luminaria sumergible RGB de 12W con control remoto. Cuerpo de acero inoxidable 316L.",
			BasePrice: 45000, Category: "luminarias", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "LUM-RGB-12W", Color: "RGB", Size: "Standard", Stock: 20}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Luminaria LED Blanco Cálido 18W", Description: "Luminaria de alto brillo 18W en blanco cálido 3000K. Incluye transformador 12V y cable 2.5m.",
			BasePrice: 38000, Category: "luminarias", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "LUM-WW-18W", Color: "Blanco Cálido", Size: "Standard", Stock: 15}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Luminaria Osire Pro 24W", Description: "Línea premium Osire RGBW 24W. Control por app, programación de escenas. Compatible con domótica.",
			BasePrice: 89000, Category: "osire", Brand: "Pooled Osire",
			Images:    []string{},
			Variants:  []Variant{{SKU: "OSR-PRO-24W", Color: "RGBW", Size: "Standard", Stock: 8}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Osire Mini 9W", Description: "Versión compacta de la línea Osire. Control por app iOS/Android. Consumo ultra eficiente.",
			BasePrice: 62000, Category: "osire", Brand: "Pooled Osire",
			Images:    []string{},
			Variants:  []Variant{{SKU: "OSR-MINI-9W", Color: "RGBW", Size: "Mini", Stock: 12}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Controlador RF 4 Canales", Description: "Controlador RF inalámbrico para hasta 4 luminarias. Alcance 30m.",
			BasePrice: 22000, Category: "controladores", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "CTRL-RF-4CH", Color: "Negro", Size: "Standard", Stock: 25}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Controlador WiFi Smart", Description: "Control desde el celular. Compatible con Google Home y Alexa.",
			BasePrice: 35000, Category: "controladores", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "CTRL-WIFI-SM", Color: "Blanco", Size: "Standard", Stock: 3}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Transformador 12V 60W", Description: "Transformador clase III para iluminación subacuática. Sellado IP67, apto intemperie.",
			BasePrice: 18000, Category: "accesorios", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "TRANS-12V-60W", Color: "Gris", Size: "60W", Stock: 30}},
			CreatedAt: now, UpdatedAt: now,
		},
		Product{
			Name: "Cable Sumergible 10m", Description: "Cable doble aislación resistente a UV y químicos. Apto 12V/24V.",
			BasePrice: 9500, Category: "accesorios", Brand: "Pooled",
			Images:    []string{},
			Variants:  []Variant{{SKU: "CAB-SUM-10M", Color: "Blanco", Size: "10m", Stock: 50}},
			CreatedAt: now, UpdatedAt: now,
		},
	}

	prodResult, err := db.Collection("products").InsertMany(ctx, products)
	if err != nil {
		log.Fatalf("error insertando productos: %v", err)
	}
	log.Printf("✓ %d productos insertados", len(prodResult.InsertedIDs))

	// ── Kits ─────────────────────────────────────────────────────────────────
	kits := []interface{}{
		Kit{
			Name: "Kit Starter — Pileta Chica", Slug: "kit-starter-chica",
			Description: "Todo lo necesario para piletas de hasta 30m³. Luminaria RGB, transformador y cable.",
			Price: 72000, OriginalPrice: ptr(85000), PoolSize: "chica", Featured: true,
			CreatedAt: now, UpdatedAt: now,
		},
		Kit{
			Name: "Kit Pro — Pileta Mediana", Slug: "kit-pro-mediana",
			Description: "Ideal para piletas de 30 a 60m³. Dos luminarias RGB con controlador RF.",
			Price: 138000, OriginalPrice: ptr(162000), PoolSize: "mediana", Featured: true,
			CreatedAt: now, UpdatedAt: now,
		},
		Kit{
			Name: "Kit Premium — Pileta Grande", Slug: "kit-premium-grande",
			Description: "Para piletas de más de 60m³. Cuatro luminarias Osire Pro con controlador WiFi.",
			Price: 398000, OriginalPrice: ptr(460000), PoolSize: "grande", Featured: true,
			CreatedAt: now, UpdatedAt: now,
		},
	}

	kitsResult, err := db.Collection("kits").InsertMany(ctx, kits)
	if err != nil {
		log.Fatalf("error insertando kits: %v", err)
	}
	log.Printf("✓ %d kits insertados", len(kitsResult.InsertedIDs))
	log.Println("Seed completado exitosamente 🎉")
}
