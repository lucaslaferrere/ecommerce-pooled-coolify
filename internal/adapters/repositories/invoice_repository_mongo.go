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

type InvoiceRepositoryMongo struct {
	collection *mongo.Collection
}

func NewInvoiceRepositoryMongo(col *mongo.Collection) *InvoiceRepositoryMongo {
	return &InvoiceRepositoryMongo{collection: col}
}

// Upsert reemplaza la factura del pedido (una por order_id).
func (r *InvoiceRepositoryMongo) Upsert(ctx context.Context, inv *domain.Invoice) error {
	now := time.Now()
	inv.UpdatedAt = now
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"order_id": inv.OrderID},
		bson.M{
			"$set": bson.M{
				"filename":   inv.Filename,
				"content":    inv.Content,
				"sent_at":    inv.SentAt,
				"updated_at": now,
			},
			"$setOnInsert": bson.M{"order_id": inv.OrderID, "created_at": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *InvoiceRepositoryMongo) GetByOrderID(ctx context.Context, orderID primitive.ObjectID) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&inv)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &inv, err
}

// SetSentAt marca la fecha de envío de la factura del pedido.
func (r *InvoiceRepositoryMongo) SetSentAt(ctx context.Context, orderID primitive.ObjectID, t time.Time) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"order_id": orderID},
		bson.M{"$set": bson.M{"sent_at": t, "updated_at": time.Now()}},
	)
	return err
}
