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

// UserRepositoryMongo es la implementación de UserRepository con MongoDB
type UserRepositoryMongo struct {
	collection *mongo.Collection
}

// NewUserRepositoryMongo crea una nueva instancia de UserRepositoryMongo
func NewUserRepositoryMongo(collection *mongo.Collection) *UserRepositoryMongo {
	return &UserRepositoryMongo{
		collection: collection,
	}
}

// Create crea un nuevo usuario en la base de datos
func (r *UserRepositoryMongo) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("usuario no puede ser nil")
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("el email ya está registrado")
		}
		return err
	}

	// Asignar el ID generado por MongoDB
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID obtiene un usuario por su ID
func (r *UserRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	if id.IsZero() {
		return nil, errors.New("ID no puede ser vacío")
	}

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Retornar nil sin error si no existe
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail obtiene un usuario por su correo electrónico
func (r *UserRepositoryMongo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, errors.New("email no puede ser vacío")
	}

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Retornar nil sin error si no existe
		}
		return nil, err
	}

	return &user, nil
}

// Update actualiza un usuario existente
func (r *UserRepositoryMongo) Update(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("usuario no puede ser nil")
	}

	if user.ID.IsZero() {
		return errors.New("ID del usuario no puede ser vacío")
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"email":         user.Email,
			"password_hash": user.PasswordHash,
			"role":          user.Role,
			"updated_at":    user.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("usuario no encontrado")
	}

	return nil
}

// Delete elimina un usuario por su ID
func (r *UserRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	if id.IsZero() {
		return errors.New("ID no puede ser vacío")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("usuario no encontrado")
	}

	return nil
}

// List obtiene una lista de usuarios con paginación
func (r *UserRepositoryMongo) List(ctx context.Context, skip int64, limit int64) ([]*domain.User, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
