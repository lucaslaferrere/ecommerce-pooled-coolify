package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository define el contrato para las operaciones de persistencia de usuarios
type UserRepository interface {
	// Create crea un nuevo usuario en la base de datos
	Create(ctx context.Context, user *domain.User) error

	// GetByID obtiene un usuario por su ID
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)

	// GetByEmail obtiene un usuario por su correo electrónico
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update actualiza un usuario existente
	Update(ctx context.Context, user *domain.User) error

	// Delete elimina un usuario por su ID
	Delete(ctx context.Context, id primitive.ObjectID) error

	// List obtiene una lista de usuarios con paginación
	List(ctx context.Context, skip int64, limit int64) ([]*domain.User, error)
}
