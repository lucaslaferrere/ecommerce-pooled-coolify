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
	ErrCouponNotFound = errors.New("cupón no encontrado")
	ErrCouponInactive = errors.New("cupón inactivo")
	ErrCouponExpired  = errors.New("cupón vencido")
)

type CouponService struct {
	repo *repositories.CouponRepositoryMongo
}

func NewCouponService(repo *repositories.CouponRepositoryMongo) *CouponService {
	return &CouponService{repo: repo}
}

// Validate checks a coupon code and returns the coupon if valid.
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
