package services

import (
	"context"

	"ecommerce-pooled/internal/adapters/repositories"
)

const abandonedCartKey = "abandoned_cart_email"

type SettingService struct {
	repo *repositories.SettingRepositoryMongo
}

func NewSettingService(repo *repositories.SettingRepositoryMongo) *SettingService {
	return &SettingService{repo: repo}
}

// DefaultAbandonedCartBody es el cuerpo por defecto del email de recuperación.
// Variables disponibles: {{codigo}}, {{descuento}}, {{vencimiento}}, {{productos}}, {{total}}.
func (s *SettingService) DefaultAbandonedCartBody() string {
	return "¡Hola! Vimos que dejaste algunos productos en tu carrito:\n\n{{productos}}\n\n" +
		"Total: {{total}}\n\nQueremos ayudarte a completar tu compra. " +
		"Usá el cupón {{codigo}} y obtené {{descuento}} de descuento. " +
		"¡Apurate que vence el {{vencimiento}}!"
}

func (s *SettingService) GetAbandonedCartBody(ctx context.Context) (string, error) {
	body, err := s.repo.Get(ctx, abandonedCartKey)
	if err != nil {
		return "", err
	}
	if body == "" {
		return s.DefaultAbandonedCartBody(), nil
	}
	return body, nil
}

func (s *SettingService) SetAbandonedCartBody(ctx context.Context, body string) error {
	return s.repo.Set(ctx, abandonedCartKey, body)
}
