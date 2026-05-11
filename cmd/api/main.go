package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-pooled/internal/adapters/database"
	"ecommerce-pooled/internal/app"
	"ecommerce-pooled/internal/config"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
)

func main() {
	// ── Configuración ────────────────────────────────────────────────────────
	_ = godotenv.Load() // carga .env si existe; en producción las vars vienen del entorno
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	log.Printf("Iniciando servidor | modo=%s puerto=%s db=%s", cfg.GinMode, cfg.Port, cfg.DBName)

	// ── Conexión a MongoDB ────────────────────────────────────────────────────
	logMongoURI(cfg.MongoURI)
	mongoClient, err := database.NewClient(context.Background(), cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	defer mongoClient.Close(context.Background())

	db := mongoClient.GetDatabase()

	// ── Índices ───────────────────────────────────────────────────────────────
	if err := ensureIndexes(context.Background(), db); err != nil {
		log.Fatalf("Error creando índices: %v", err)
	}

	// ── Router (wiring completo en internal/app) ──────────────────────────────
	router := app.BuildRouter(cfg, db)

	// ── Servidor HTTP con graceful shutdown ───────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Servidor escuchando en http://0.0.0.0:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error al iniciar servidor: %v", err)
		}
	}()

	// Esperar SIGINT / SIGTERM (Ctrl+C o docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Señal de apagado recibida, drenando peticiones activas...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Graceful shutdown forzado: %v", err)
	}
	log.Println("Servidor apagado correctamente")
}

// logMongoURI imprime la URI con la contraseña enmascarada.
// Si la URI no contiene usuario, imprime una advertencia de credenciales ausentes.
func logMongoURI(rawURI string) {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		log.Printf("[MONGO] URI inválida: %v", err)
		return
	}
	if parsed.User == nil || parsed.User.Username() == "" {
		log.Printf("[MONGO] ADVERTENCIA: URI cargada sin credenciales → %s (verifica tu .env)", rawURI)
		return
	}
	safe := *parsed
	safe.User = url.UserPassword(parsed.User.Username(), "***")
	log.Printf("[MONGO] URI cargada → %s", safe.String())
}

// ensureIndexes crea los índices de MongoDB al arrancar (idempotente).
// Los códigos 85/86 significan "índice ya existe" y se ignoran.
func ensureIndexes(ctx context.Context, db *mongo.Database) error {
	type spec struct {
		coll  string
		model mongo.IndexModel
	}
	specs := []spec{
		{
			coll: "users",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("email_unique"),
			},
		},
		{
			coll: "orders",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("orders_user_id"),
			},
		},
		{
			coll: "orders",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "status", Value: 1}},
				Options: options.Index().SetName("orders_status"),
			},
		},
		{
			coll: "products",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "variants.sku", Value: 1}},
				Options: options.Index().SetName("products_variant_sku"),
			},
		},
	}

	for _, s := range specs {
		if _, err := db.Collection(s.coll).Indexes().CreateOne(ctx, s.model); err != nil {
			var cmdErr mongo.CommandError
			if errors.As(err, &cmdErr) && (cmdErr.Code == 85 || cmdErr.Code == 86) {
				continue
			}
			return err
		}
	}
	return nil
}
