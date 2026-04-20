package ports

import (
	"context"

	"ecommerce-pooled/internal/core/domain"
)

type DistributorLeadRepository interface {
	Create(ctx context.Context, lead *domain.DistributorLead) error
}
