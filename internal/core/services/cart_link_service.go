package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"ecommerce-pooled/internal/core/domain"
)

type CartLinkRepository interface {
	Create(ctx context.Context, link *domain.CartLink) error
	GetByToken(ctx context.Context, token string) (*domain.CartLink, error)
}

type CartLinkService struct {
	repo CartLinkRepository
}

func NewCartLinkService(repo CartLinkRepository) *CartLinkService {
	return &CartLinkService{repo: repo}
}

func (s *CartLinkService) Create(ctx context.Context, items []domain.CartLinkItem) (*domain.CartLink, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	link := &domain.CartLink{
		Token:     hex.EncodeToString(b),
		Items:     items,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(72 * time.Hour), // expira en 3 días
	}
	if err := s.repo.Create(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *CartLinkService) GetByToken(ctx context.Context, token string) (*domain.CartLink, error) {
	return s.repo.GetByToken(ctx, token)
}
