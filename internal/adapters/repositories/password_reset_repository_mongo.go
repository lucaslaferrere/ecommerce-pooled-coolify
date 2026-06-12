package repositories

import (
	"context"
	"errors"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PasswordResetRepositoryMongo implementa PasswordResetRepository con MongoDB
type PasswordResetRepositoryMongo struct {
	collection *mongo.Collection
}

// NewPasswordResetRepositoryMongo crea una nueva instancia
func NewPasswordResetRepositoryMongo(collection *mongo.Collection) *PasswordResetRepositoryMongo {
	return &PasswordResetRepositoryMongo{collection: collection}
}

// Create inserta un nuevo token de reset en la colección
func (r *PasswordResetRepositoryMongo) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	if token == nil {
		return errors.New("token no puede ser nil")
	}
	result, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	token.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetActiveByEmailAndCode busca un token no usado y no vencido para email+code
func (r *PasswordResetRepositoryMongo) GetActiveByEmailAndCode(ctx context.Context, email, code string) (*domain.PasswordResetToken, error) {
	filter := bson.M{
		"email":      email,
		"code":       code,
		"used":       false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var token domain.PasswordResetToken
	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// MarkAsUsed marca el token como consumido para que no pueda reutilizarse
func (r *PasswordResetRepositoryMongo) MarkAsUsed(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"used": true}},
	)
	return err
}

// InvalidateByEmail pone used=true en todos los tokens pendientes del email
func (r *PasswordResetRepositoryMongo) InvalidateByEmail(ctx context.Context, email string) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"email": email, "used": false},
		bson.M{"$set": bson.M{"used": true}},
	)
	return err
}
