package repositories

import (
	"context"
	"strings"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CouponRepositoryMongo struct {
	collection *mongo.Collection
}

func NewCouponRepositoryMongo(col *mongo.Collection) *CouponRepositoryMongo {
	return &CouponRepositoryMongo{collection: col}
}

func (r *CouponRepositoryMongo) FindByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var c domain.Coupon
	err := r.collection.FindOne(ctx, bson.M{"code": strings.ToUpper(code)}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

func (r *CouponRepositoryMongo) List(ctx context.Context) ([]domain.Coupon, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Coupon
	return out, cur.All(ctx, &out)
}

func (r *CouponRepositoryMongo) Create(ctx context.Context, c *domain.Coupon) error {
	c.Code = strings.ToUpper(c.Code)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CouponRepositoryMongo) Update(ctx context.Context, c *domain.Coupon) error {
	c.Code = strings.ToUpper(c.Code)
	c.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": c.ID}, c)
	return err
}

func (r *CouponRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
