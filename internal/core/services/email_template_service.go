package services

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailTemplateService struct {
	repo       *repositories.EmailTemplateRepositoryMongo
	settingSvc *SettingService
}

func NewEmailTemplateService(repo *repositories.EmailTemplateRepositoryMongo, settingSvc *SettingService) *EmailTemplateService {
	return &EmailTemplateService{repo: repo, settingSvc: settingSvc}
}

// ensureSeed creates one initial template from the existing global body the
// first time the collection is empty, so the current text is not lost.
func (s *EmailTemplateService) ensureSeed(ctx context.Context) error {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	body, err := s.settingSvc.GetAbandonedCartBody(ctx)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &domain.EmailTemplate{Name: "Predeterminada", Body: body})
}

func (s *EmailTemplateService) List(ctx context.Context) ([]domain.EmailTemplate, error) {
	if err := s.ensureSeed(ctx); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}

func (s *EmailTemplateService) Create(ctx context.Context, name, body string) (*domain.EmailTemplate, error) {
	if name == "" || body == "" {
		return nil, errors.New("name y body son requeridos")
	}
	t := &domain.EmailTemplate{Name: name, Body: body}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *EmailTemplateService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *EmailTemplateService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.EmailTemplate, error) {
	return s.repo.GetByID(ctx, id)
}
