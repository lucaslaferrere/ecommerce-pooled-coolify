package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PasswordResetRepository define las operaciones de persistencia para tokens de reset
type PasswordResetRepository interface {
	// Create guarda un nuevo token de reset
	Create(ctx context.Context, token *domain.PasswordResetToken) error

	// GetActiveByEmailAndCode obtiene un token válido (no usado, no vencido)
	GetActiveByEmailAndCode(ctx context.Context, email, code string) (*domain.PasswordResetToken, error)

	// MarkAsUsed marca un token como consumido
	MarkAsUsed(ctx context.Context, id primitive.ObjectID) error

	// InvalidateByEmail invalida todos los tokens pendientes de un email
	InvalidateByEmail(ctx context.Context, email string) error
}
