package repositories

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type KitRepositoryMongo struct {
	collection *mongo.Collection
}

func NewKitRepositoryMongo(collection *mongo.Collection) *KitRepositoryMongo {
	return &KitRepositoryMongo{collection: collection}
}

func (r *KitRepositoryMongo) Create(ctx context.Context, kit *domain.Kit) error {
	result, err := r.collection.InsertOne(ctx, kit)
	if err != nil {
		return err
	}
	kit.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *KitRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Kit, error) {
	var kit domain.Kit
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&kit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &kit, nil
}

func (r *KitRepositoryMongo) List(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}
	return kits, nil
}

func (r *KitRepositoryMongo) ListFeatured(ctx context.Context) ([]*domain.Kit, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"featured": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}
	return kits, nil
}

func (r *KitRepositoryMongo) Update(ctx context.Context, kit *domain.Kit) error {
	filter := bson.M{"_id": kit.ID}
	update := bson.M{"$set": kit}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("kit no encontrado")
	}
	return nil
}

func (r *KitRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("kit no encontrado")
	}
	return nil
}
