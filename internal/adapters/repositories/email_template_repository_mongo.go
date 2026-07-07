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

type EmailTemplateRepositoryMongo struct {
	collection *mongo.Collection
}

func NewEmailTemplateRepositoryMongo(col *mongo.Collection) *EmailTemplateRepositoryMongo {
	return &EmailTemplateRepositoryMongo{collection: col}
}

func (r *EmailTemplateRepositoryMongo) List(ctx context.Context) ([]domain.EmailTemplate, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.EmailTemplate
	return out, cur.All(ctx, &out)
}

func (r *EmailTemplateRepositoryMongo) Create(ctx context.Context, t *domain.EmailTemplate) error {
	t.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, t)
	if err != nil {
		return err
	}
	t.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *EmailTemplateRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *EmailTemplateRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.EmailTemplate, error) {
	var t domain.EmailTemplate
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &t, err
}

func (r *EmailTemplateRepositoryMongo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
