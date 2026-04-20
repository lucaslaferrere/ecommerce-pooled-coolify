package services

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"
)

type DistributorLeadService struct {
	repo ports.DistributorLeadRepository
}

func NewDistributorLeadService(repo ports.DistributorLeadRepository) *DistributorLeadService {
	return &DistributorLeadService{repo: repo}
}

func (s *DistributorLeadService) CreateLead(ctx context.Context, lead *domain.DistributorLead) error {
	if lead.FullName == "" {
		return errors.New("nombre requerido")
	}
	if lead.Email == "" {
		return errors.New("email requerido")
	}
	if lead.Company == "" {
		return errors.New("empresa requerida")
	}
	lead.Status = "new"
	return s.repo.Create(ctx, lead)
}
