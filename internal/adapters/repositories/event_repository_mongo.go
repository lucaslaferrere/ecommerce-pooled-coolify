package repositories

import (
	"context"
	"fmt"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepositoryMongo struct {
	collection *mongo.Collection
}

func NewEventRepositoryMongo(collection *mongo.Collection) *EventRepositoryMongo {
	return &EventRepositoryMongo{collection: collection}
}

func (r *EventRepositoryMongo) Create(ctx context.Context, event *domain.Event) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *EventRepositoryMongo) CountByType(ctx context.Context, eventType string, from, to time.Time) (int64, error) {
	filter := bson.M{
		"type":       eventType,
		"created_at": bson.M{"$gte": from, "$lte": to},
	}
	return r.collection.CountDocuments(ctx, filter)
}

func (r *EventRepositoryMongo) GroupByDay(ctx context.Context, eventType string, from, to time.Time) ([]domain.DayCount, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"type":       eventType,
			"created_at": bson.M{"$gte": from, "$lte": to},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format":   "%Y-%m-%d",
					"date":     "$created_at",
					"timezone": "America/Argentina/Buenos_Aires",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []domain.DayCount
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *EventRepositoryMongo) TopItems(ctx context.Context, eventType, idField, nameField string, limit int, from, to time.Time) ([]domain.ItemStat, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"type":                                      eventType,
			"created_at":                                bson.M{"$gte": from, "$lte": to},
			fmt.Sprintf("payload.%s", idField):          bson.M{"$exists": true, "$ne": ""},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   fmt.Sprintf("$payload.%s", idField),
			"name":  bson.M{"$first": fmt.Sprintf("$payload.%s", nameField)},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []domain.ItemStat
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *EventRepositoryMongo) CountUniqueSessions(ctx context.Context, eventType string, from, to time.Time) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"type":       eventType,
			"created_at": bson.M{"$gte": from, "$lte": to},
			"session_id": bson.M{"$ne": ""},
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$session_id"}}},
		{{Key: "$count", Value: "count"}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Count int64 `bson:"count"`
	}
	if err = cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0, err
	}
	return result[0].Count, nil
}
