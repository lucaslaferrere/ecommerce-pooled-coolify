package services

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
)

type WizardRecommendationService struct {
	repo ports.WizardRecommendationRepository
}

func NewWizardRecommendationService(repo ports.WizardRecommendationRepository) *WizardRecommendationService {
	return &WizardRecommendationService{repo: repo}
}

func (s *WizardRecommendationService) LogRecommendation(ctx context.Context, rec *domain.WizardRecommendation) error {
	return s.repo.Create(ctx, rec)
}
