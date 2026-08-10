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

type AuditLogRepositoryMongo struct {
	collection *mongo.Collection
}

func NewAuditLogRepositoryMongo(col *mongo.Collection) *AuditLogRepositoryMongo {
	return &AuditLogRepositoryMongo{collection: col}
}

func (r *AuditLogRepositoryMongo) Create(ctx context.Context, entry *domain.AdminAuditLog) error {
	entry.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, entry)
	return err
}

// List devuelve entradas más recientes primero, con filtros opcionales por
// admin (adminID no-zero) y rango de fecha (from/to, cero = sin límite).
func (r *AuditLogRepositoryMongo) List(ctx context.Context, adminID primitive.ObjectID, from, to time.Time, skip, limit int64) ([]*domain.AdminAuditLog, error) {
	filter := bson.M{}
	if !adminID.IsZero() {
		filter["admin_id"] = adminID
	}
	dateFilter := bson.M{}
	if !from.IsZero() {
		dateFilter["$gte"] = from
	}
	if !to.IsZero() {
		dateFilter["$lte"] = to
	}
	if len(dateFilter) > 0 {
		filter["created_at"] = dateFilter
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(skip).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	logs := []*domain.AdminAuditLog{}
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditLogRepositoryMongo) Count(ctx context.Context, adminID primitive.ObjectID, from, to time.Time) (int64, error) {
	filter := bson.M{}
	if !adminID.IsZero() {
		filter["admin_id"] = adminID
	}
	dateFilter := bson.M{}
	if !from.IsZero() {
		dateFilter["$gte"] = from
	}
	if !to.IsZero() {
		dateFilter["$lte"] = to
	}
	if len(dateFilter) > 0 {
		filter["created_at"] = dateFilter
	}
	return r.collection.CountDocuments(ctx, filter)
}

// DistinctAdmins devuelve los admins que tienen al menos una entrada en el
// log, para poblar el filtro del frontend.
func (r *AuditLogRepositoryMongo) DistinctAdmins(ctx context.Context) ([]struct {
	AdminID    primitive.ObjectID `bson:"_id"`
	AdminEmail string             `bson:"admin_email"`
}, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":         "$admin_id",
			"admin_email": bson.M{"$first": "$admin_email"},
		}}},
		{{Key: "$sort", Value: bson.M{"admin_email": 1}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var out []struct {
		AdminID    primitive.ObjectID `bson:"_id"`
		AdminEmail string             `bson:"admin_email"`
	}
	return out, cursor.All(ctx, &out)
}
