package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KitRepository interface {
	Create(ctx context.Context, kit *domain.Kit) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Kit, error)
	List(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error)
	ListFeatured(ctx context.Context) ([]*domain.Kit, error)
	Update(ctx context.Context, kit *domain.Kit) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
