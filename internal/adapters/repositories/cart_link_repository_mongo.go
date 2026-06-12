package repositories

import (
	"context"
	"errors"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartLinkRepositoryMongo struct {
	collection *mongo.Collection
}

func NewCartLinkRepositoryMongo(col *mongo.Collection) *CartLinkRepositoryMongo {
	// TTL index: MongoDB borra documentos automáticamente cuando expiran
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"expires_at": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"token": 1},
		Options: options.Index().SetUnique(true),
	})
	return &CartLinkRepositoryMongo{collection: col}
}

func (r *CartLinkRepositoryMongo) Create(ctx context.Context, link *domain.CartLink) error {
	_, err := r.collection.InsertOne(ctx, link)
	return err
}

func (r *CartLinkRepositoryMongo) GetByToken(ctx context.Context, token string) (*domain.CartLink, error) {
	var link domain.CartLink
	err := r.collection.FindOne(ctx, bson.M{
		"token":      token,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&link)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("link no encontrado o expirado")
	}
	return &link, err
}
