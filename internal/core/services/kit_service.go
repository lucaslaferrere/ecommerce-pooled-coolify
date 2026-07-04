package services

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

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

func (s *KitService) ListVisibleKits(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.kitRepository.ListVisible(ctx, skip, limit)
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

// BulkUpdatePrice aplica un porcentaje al precio de un conjunto de kits.
// Si ids no está vacío, esos kits; si no, todos. Ajusta Price y, si existe,
// OriginalPrice (para preservar el descuento visual). Devuelve cuántos actualizó.
func (s *KitService) BulkUpdatePrice(ctx context.Context, percent float64, ids []primitive.ObjectID) (int, error) {
	if percent == 0 {
		return 0, errors.New("el porcentaje no puede ser 0")
	}
	if percent < -90 || percent > 1000 {
		return 0, errors.New("porcentaje fuera de rango (-90 a 1000)")
	}

	var kits []*domain.Kit
	if len(ids) > 0 {
		for _, id := range ids {
			k, err := s.kitRepository.GetByID(ctx, id)
			if err != nil {
				log.Printf("BulkUpdatePrice(kits): error obteniendo kit %s: %v", id.Hex(), err)
				continue
			}
			if k == nil {
				continue
			}
			kits = append(kits, k)
		}
	} else {
		list, err := s.kitRepository.List(ctx, 0, 100000)
		if err != nil {
			return 0, err
		}
		kits = list
	}

	factor := 1 + percent/100
	updated := 0
	for _, k := range kits {
		newPrice := math.Round(k.Price * factor)
		if newPrice < 0 {
			newPrice = 0
		}
		k.Price = newPrice
		if k.OriginalPrice != nil {
			newOriginal := math.Round(*k.OriginalPrice * factor)
			if newOriginal < 0 {
				newOriginal = 0
			}
			k.OriginalPrice = &newOriginal
		}
		k.UpdatedAt = time.Now()
		if err := s.kitRepository.Update(ctx, k); err != nil {
			log.Printf("BulkUpdatePrice(kits): no se pudo actualizar kit %s: %v", k.ID.Hex(), err)
			continue
		}
		updated++
	}
	return updated, nil
}
