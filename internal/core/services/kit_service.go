package services

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KitService struct {
	kitRepository ports.KitRepository
}

func NewKitService(kitRepository ports.KitRepository) *KitService {
	return &KitService{kitRepository: kitRepository}
}

func (s *KitService) CreateKit(ctx context.Context, kit *domain.Kit) error {
	if kit.Name == "" {
		return errors.New("nombre de kit requerido")
	}
	if kit.Price <= 0 {
		return errors.New("precio debe ser mayor a 0")
	}
	return s.kitRepository.Create(ctx, kit)
}

func (s *KitService) GetKit(ctx context.Context, id primitive.ObjectID) (*domain.Kit, error) {
	kit, err := s.kitRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if kit == nil {
		return nil, errors.New("kit no encontrado")
	}
	return kit, nil
}

func (s *KitService) ListKits(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.kitRepository.List(ctx, skip, limit)
}

func (s *KitService) ListFeaturedKits(ctx context.Context) ([]*domain.Kit, error) {
	return s.kitRepository.ListFeatured(ctx)
}

func (s *KitService) UpdateKit(ctx context.Context, kit *domain.Kit) error {
	if kit.Name == "" {
		return errors.New("nombre de kit requerido")
	}
	return s.kitRepository.Update(ctx, kit)
}

func (s *KitService) DeleteKit(ctx context.Context, id primitive.ObjectID) error {
	return s.kitRepository.Delete(ctx, id)
}
