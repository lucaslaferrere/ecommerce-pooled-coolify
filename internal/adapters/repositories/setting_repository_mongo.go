package repositories

import (
	"context"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingRepositoryMongo struct {
	collection *mongo.Collection
}

func NewSettingRepositoryMongo(col *mongo.Collection) *SettingRepositoryMongo {
	return &SettingRepositoryMongo{collection: col}
}

// Get devuelve el value de una key, o "" si no existe.
func (r *SettingRepositoryMongo) Get(ctx context.Context, key string) (string, error) {
	var s domain.Setting
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	return s.Value, err
}

func (r *SettingRepositoryMongo) Set(ctx context.Context, key, value string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"key": key},
		bson.M{"$set": bson.M{"value": value}},
		options.Update().SetUpsert(true),
	)
	return err
}
