package services

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCouponNotFound       = errors.New("cupón no encontrado")
	ErrCouponInactive       = errors.New("cupón inactivo")
	ErrCouponExpired        = errors.New("cupón vencido")
	ErrCouponMaxUses        = errors.New("cupón agotado")
	ErrCouponMaxUsesPerUser = errors.New("ya usaste este cupón")
)

// CouponUsageCounter es el puerto que el servicio usa para conocer cuántas veces
// se redimió un cupón en órdenes pagadas. El repositorio de órdenes lo satisface.
type CouponUsageCounter interface {
	CountPaidByCoupon(ctx context.Context, code string) (int64, error)
	CountPaidByCouponAndUser(ctx context.Context, code string, userID primitive.ObjectID) (int64, error)
}

type CouponService struct {
	repo    *repositories.CouponRepositoryMongo
	counter CouponUsageCounter
}

func NewCouponService(repo *repositories.CouponRepositoryMongo, counter CouponUsageCounter) *CouponService {
	return &CouponService{repo: repo, counter: counter}
}

// Validate checks a coupon code and returns the coupon if active and not expired.
// No verifica límites de uso — para eso usar ValidateForUser.
func (s *CouponService) Validate(ctx context.Context, code string) (*domain.Coupon, error) {
	c, err := s.repo.FindByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCouponNotFound
	}
	if !c.Active {
		return nil, ErrCouponInactive
	}
	if c.ExpiresAt != nil && time.Now().After(*c.ExpiresAt) {
		return nil, ErrCouponExpired
	}
	return c, nil
}

// ValidateForUser valida el cupón y además verifica los límites de uso contra las
// órdenes pagadas. Si userID es cero (validación pública sin sesión) solo se
// verifica el límite total; el límite por cliente se hace cumplir en el checkout.
func (s *CouponService) ValidateForUser(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Coupon, error) {
	c, err := s.Validate(ctx, code)
	if err != nil {
		return nil, err
	}
	if c.MaxUses > 0 {
		total, err := s.counter.CountPaidByCoupon(ctx, c.Code)
		if err != nil {
			return nil, err
		}
		if total >= int64(c.MaxUses) {
			return nil, ErrCouponMaxUses
		}
	}
	if c.MaxUsesPerUser > 0 && !userID.IsZero() {
		used, err := s.counter.CountPaidByCouponAndUser(ctx, c.Code, userID)
		if err != nil {
			return nil, err
		}
		if used >= int64(c.MaxUsesPerUser) {
			return nil, ErrCouponMaxUsesPerUser
		}
	}
	return c, nil
}

// ApplyDiscount returns the discount amount for a given subtotal and coupon.
func ApplyCouponDiscount(subtotal, discountPercent float64) float64 {
	return math.Round(subtotal * discountPercent / 100)
}

func (s *CouponService) List(ctx context.Context) ([]domain.Coupon, error) {
	return s.repo.List(ctx)
}

func (s *CouponService) Create(ctx context.Context, c *domain.Coupon) error {
	return s.repo.Create(ctx, c)
}

func (s *CouponService) Update(ctx context.Context, c *domain.Coupon) error {
	return s.repo.Update(ctx, c)
}

func (s *CouponService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}
