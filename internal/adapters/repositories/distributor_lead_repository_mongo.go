package repositories

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DistributorLeadRepositoryMongo struct {
	collection *mongo.Collection
}

func NewDistributorLeadRepositoryMongo(collection *mongo.Collection) *DistributorLeadRepositoryMongo {
	return &DistributorLeadRepositoryMongo{collection: collection}
}

func (r *DistributorLeadRepositoryMongo) Create(ctx context.Context, lead *domain.DistributorLead) error {
	result, err := r.collection.InsertOne(ctx, lead)
	if err != nil {
		return err
	}
	lead.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
