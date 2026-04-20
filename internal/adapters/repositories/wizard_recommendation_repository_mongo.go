package repositories

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WizardRecommendationRepositoryMongo struct {
	collection *mongo.Collection
}

func NewWizardRecommendationRepositoryMongo(collection *mongo.Collection) *WizardRecommendationRepositoryMongo {
	return &WizardRecommendationRepositoryMongo{collection: collection}
}

func (r *WizardRecommendationRepositoryMongo) Create(ctx context.Context, rec *domain.WizardRecommendation) error {
	result, err := r.collection.InsertOne(ctx, rec)
	if err != nil {
		return err
	}
	rec.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
