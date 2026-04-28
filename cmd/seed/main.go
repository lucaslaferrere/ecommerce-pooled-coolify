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

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := database.NewClient(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	defer client.Close(context.Background())

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error generando bcrypt hash: %v", err)
	}

	now := time.Now()
	users := client.GetDatabase().Collection("users")

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
}
