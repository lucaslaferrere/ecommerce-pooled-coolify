package repositories

import (
	"context"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartRepositoryMongo struct {
	collection *mongo.Collection
}

func NewCartRepositoryMongo(col *mongo.Collection) *CartRepositoryMongo {
	return &CartRepositoryMongo{collection: col}
}

// Upsert reemplaza los items del carrito del usuario (crea el doc si no existe).
func (r *CartRepositoryMongo) Upsert(ctx context.Context, userID primitive.ObjectID, items []domain.CartItem) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"items": items, "updated_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *CartRepositoryMongo) DeleteByUser(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

func (r *CartRepositoryMongo) GetByUser(ctx context.Context, userID primitive.ObjectID) (*domain.Cart, error) {
	var c domain.Cart
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

// ListNonEmpty devuelve los carritos que tienen al menos un item.
func (r *CartRepositoryMongo) ListNonEmpty(ctx context.Context) ([]domain.Cart, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cur, err := r.collection.Find(ctx, bson.M{"items.0": bson.M{"$exists": true}}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Cart
	return out, cur.All(ctx, &out)
}

func (r *CartRepositoryMongo) SetReminderSentAt(ctx context.Context, userID primitive.ObjectID, t time.Time) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"last_reminder_sent_at": t}},
	)
	return err
}
