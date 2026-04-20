package services

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService contiene la lógica de negocio relacionada con usuarios
type UserService struct {
	userRepository ports.UserRepository
}

// NewUserService crea una nueva instancia de UserService
func NewUserService(userRepository ports.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

// CreateUser crea un nuevo usuario
// Esta es la lógica de negocio pura que puede ser testeada fácilmente
func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	// Validaciones de negocio
	if user.Email == "" {
		return ErrInvalidEmail
	}

	if len(user.PasswordHash) == 0 {
		return ErrInvalidPassword
	}

	// Verificar que el email no exista
	existingUser, err := s.userRepository.GetByEmail(ctx, user.Email)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if existingUser != nil {
		return ErrEmailAlreadyExists
	}

	// Crear el usuario
	return s.userRepository.Create(ctx, user)
}

// GetUser obtiene un usuario por ID
func (s *UserService) GetUser(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	return s.userRepository.GetByID(ctx, id)
}

// GetUserByEmail obtiene un usuario por email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepository.GetByEmail(ctx, email)
}

// UpdateUser actualiza un usuario
func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if user.Email == "" {
		return ErrInvalidEmail
	}

	return s.userRepository.Update(ctx, user)
}

// DeleteUser elimina un usuario
func (s *UserService) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	return s.userRepository.Delete(ctx, id)
}

// ListUsers obtiene una lista de usuarios con paginación
func (s *UserService) ListUsers(ctx context.Context, skip int64, limit int64) ([]*domain.User, error) {
	return s.userRepository.List(ctx, skip, limit)
}
