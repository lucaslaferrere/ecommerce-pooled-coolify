// Seed inserta productos y kits de ejemplo para AquaLed.
// Uso: go run cmd/seed/main.go
package main

import (
	"context"
	"log"
	"time"

	"ecommerce-pooled/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ProductIDs    []string           `bson:"product_ids"`
	Featured      bool               `bson:"featured"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

func ptr(f float64) *float64 { return &f }

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("error conectando a MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(cfg.DBName)
	now := time.Now()

	// ── Limpiar colecciones ───────────────────────────────────────────────────
	db.Collection("products").DeleteMany(ctx, bson.M{})
	db.Collection("kits").DeleteMany(ctx, bson.M{})
	log.Println("Colecciones limpiadas")

	// ── Productos ─────────────────────────────────────────────────────────────
	products := []interface{}{
		Product{
			Name:        "Luminaria LED RGB 12W",
			Description: "Luminaria sumergible RGB de 12W con control por control remoto. Cuerpo de acero inoxidable 316L, resistente a químicos de pileta. Ideal para piletas de fibra y hormigón.",
			BasePrice:   45000,
			Category:    "luminarias",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "LUM-RGB-12W", Color: "RGB", Size: "Standard", Stock: 20, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Luminaria LED Blanco Cálido 18W",
			Description: "Luminaria de alto brillo 18W en blanco cálido 3000K. Perfecta para ambientación relajante. Incluye transformador de 12V y cable de 2.5m.",
			BasePrice:   38000,
			Category:    "luminarias",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "LUM-WW-18W", Color: "Blanco Cálido", Size: "Standard", Stock: 15, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Luminaria Osire Pro 24W",
			Description: "Línea premium Osire con tecnología RGBW de 24W. Control por app, cambios de color suaves y programación de escenas. Compatible con domótica.",
			BasePrice:   89000,
			Category:    "osire",
			Brand:       "AquaLed Osire",
			Images:      []string{},
			Variants:    []Variant{{SKU: "OSR-PRO-24W", Color: "RGBW", Size: "Standard", Stock: 8, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Osire Mini 9W",
			Description: "Versión compacta de la línea Osire, ideal para nichos y piletas chicas. Control por app iOS/Android. Consumo ultra eficiente.",
			BasePrice:   62000,
			Category:    "osire",
			Brand:       "AquaLed Osire",
			Images:      []string{},
			Variants:    []Variant{{SKU: "OSR-MINI-9W", Color: "RGBW", Size: "Mini", Stock: 12, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Controlador RF 4 Canales",
			Description: "Controlador RF inalámbrico para hasta 4 luminarias independientes. Alcance 30m. Compatible con todas las luminarias AquaLed RGB.",
			BasePrice:   22000,
			Category:    "controladores",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "CTRL-RF-4CH", Color: "Negro", Size: "Standard", Stock: 25, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Controlador WiFi Smart",
			Description: "Controlador WiFi para gestión desde el celular. Compatible con Google Home y Alexa. Permite programar horarios y escenas de color.",
			BasePrice:   35000,
			Category:    "controladores",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "CTRL-WIFI-SM", Color: "Blanco", Size: "Standard", Stock: 3, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Transformador 12V 60W",
			Description: "Transformador de seguridad clase III para iluminación subacuática. Sellado IP67, apto intemperie. Incluye fusible de protección.",
			BasePrice:   18000,
			Category:    "accesorios",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "TRANS-12V-60W", Color: "Gris", Size: "60W", Stock: 30, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
		Product{
			Name:        "Cable Sumergible 10m",
			Description: "Cable especial para instalaciones subacuáticas, doble aislación, resistente a UV y químicos de pileta. Apto 12V/24V.",
			BasePrice:   9500,
			Category:    "accesorios",
			Brand:       "AquaLed",
			Images:      []string{},
			Variants:    []Variant{{SKU: "CAB-SUM-10M", Color: "Blanco", Size: "10m", Stock: 50, PriceAdjustment: 0}},
			CreatedAt:   now, UpdatedAt: now,
		},
	}

	result, err := db.Collection("products").InsertMany(ctx, products)
	if err != nil {
		log.Fatalf("error insertando productos: %v", err)
	}
	log.Printf("✓ %d productos insertados", len(result.InsertedIDs))

	// ── Kits ─────────────────────────────────────────────────────────────────
	kits := []interface{}{
		Kit{
			Name:          "Kit Starter — Pileta Chica",
			Slug:          "kit-starter-chica",
			Description:   "Todo lo que necesitás para iluminar una pileta de hasta 30m³. Incluye luminaria RGB, transformador y cable. Instalación en 1 hora.",
			Price:         72000,
			OriginalPrice: ptr(85000),
			ImageURL:      "",
			PoolSize:      "chica",
			Featured:      true,
			CreatedAt:     now, UpdatedAt: now,
		},
		Kit{
			Name:          "Kit Pro — Pileta Mediana",
			Slug:          "kit-pro-mediana",
			Description:   "Ideal para piletas de 30 a 60m³. Dos luminarias RGB sincronizadas con controlador RF. Efecto espejo de agua garantizado.",
			Price:         138000,
			OriginalPrice: ptr(162000),
			ImageURL:      "",
			PoolSize:      "mediana",
			Featured:      true,
			CreatedAt:     now, UpdatedAt: now,
		},
		Kit{
			Name:          "Kit Premium — Pileta Grande",
			Slug:          "kit-premium-grande",
			Description:   "Para piletas de más de 60m³ o uso comercial. Cuatro luminarias Osire Pro con controlador WiFi. Control total desde el celular.",
			Price:         398000,
			OriginalPrice: ptr(460000),
			ImageURL:      "",
			PoolSize:      "grande",
			Featured:      true,
			CreatedAt:     now, UpdatedAt: now,
		},
	}

	resultKits, err := db.Collection("kits").InsertMany(ctx, kits)
	if err != nil {
		log.Fatalf("error insertando kits: %v", err)
	}
	log.Printf("✓ %d kits insertados", len(resultKits.InsertedIDs))
	log.Println("Seed completado exitosamente 🎉")
}
