package services

import (
	"context"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartStorageService struct {
	repo *repositories.CartRepositoryMongo
}

func NewCartStorageService(repo *repositories.CartRepositoryMongo) *CartStorageService {
	return &CartStorageService{repo: repo}
}

func (s *CartStorageService) Save(ctx context.Context, userID primitive.ObjectID, items []domain.CartItem) error {
	return s.repo.Upsert(ctx, userID, items)
}

func (s *CartStorageService) Clear(ctx context.Context, userID primitive.ObjectID) error {
	return s.repo.DeleteByUser(ctx, userID)
}

func (s *CartStorageService) List(ctx context.Context) ([]domain.Cart, error) {
	return s.repo.ListNonEmpty(ctx)
}

func (s *CartStorageService) Get(ctx context.Context, userID primitive.ObjectID) (*domain.Cart, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *CartStorageService) MarkReminderSent(ctx context.Context, userID primitive.ObjectID) error {
	return s.repo.SetReminderSentAt(ctx, userID, time.Now())
}
