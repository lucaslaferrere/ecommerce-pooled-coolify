package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
)

type WizardRecommendationRepository interface {
	Create(ctx context.Context, rec *domain.WizardRecommendation) error
}
