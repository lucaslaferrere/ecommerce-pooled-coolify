package services

import (
	"context"
	"testing"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeKitRepo struct {
	kits    map[primitive.ObjectID]*domain.Kit
	updated []primitive.ObjectID
}

func (f *fakeKitRepo) Create(ctx context.Context, kit *domain.Kit) error { return nil }

func (f *fakeKitRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Kit, error) {
	return f.kits[id], nil
}

func (f *fakeKitRepo) List(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error) {
	out := make([]*domain.Kit, 0, len(f.kits))
	for _, k := range f.kits {
		out = append(out, k)
	}
	return out, nil
}

func (f *fakeKitRepo) ListVisible(ctx context.Context, skip int64, limit int64) ([]*domain.Kit, error) {
	return f.List(ctx, skip, limit)
}

func (f *fakeKitRepo) ListFeatured(ctx context.Context) ([]*domain.Kit, error) {
	return f.List(ctx, 0, 0)
}

func (f *fakeKitRepo) Update(ctx context.Context, kit *domain.Kit) error {
	f.kits[kit.ID] = kit
	f.updated = append(f.updated, kit.ID)
	return nil
}

func (f *fakeKitRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	delete(f.kits, id)
	return nil
}

type fakeProductRepo struct {
	products map[primitive.ObjectID]*domain.Product
}

func (f *fakeProductRepo) Create(ctx context.Context, p *domain.Product) error { return nil }

func (f *fakeProductRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	return f.products[id], nil
}

func (f *fakeProductRepo) Update(ctx context.Context, p *domain.Product) error     { return nil }
func (f *fakeProductRepo) Delete(ctx context.Context, id primitive.ObjectID) error { return nil }

func (f *fakeProductRepo) List(ctx context.Context, filter map[string]interface{}, skip int64, limit int64) ([]*domain.Product, error) {
	ids, _ := filter["_id"].(map[string]interface{})
	wanted, _ := ids["$in"].([]primitive.ObjectID)
	out := make([]*domain.Product, 0, len(wanted))
	for _, id := range wanted {
		if p, ok := f.products[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeProductRepo) GetByCategory(ctx context.Context, category string, skip int64, limit int64) ([]*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) GetByBrand(ctx context.Context, brand string, skip int64, limit int64) ([]*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) UpdateVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, newStock int) error {
	return nil
}
func (f *fakeProductRepo) DecrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error {
	return nil
}
func (f *fakeProductRepo) IncrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error {
	return nil
}
func (f *fakeProductRepo) DecrementProductStock(ctx context.Context, productID primitive.ObjectID, quantity int) error {
	return nil
}
func (f *fakeProductRepo) IncrementProductStock(ctx context.Context, productID primitive.ObjectID, quantity int) error {
	return nil
}

func TestRefreshSuggestedPrices(t *testing.T) {
	ledID := primitive.NewObjectID()
	controllerID := primitive.NewObjectID()

	kitUpToDate := &domain.Kit{
		ID:         primitive.NewObjectID(),
		Name:       "Kit al día",
		Price:      300, // 100 (led) + 200 (controller) — ya coincide
		ProductIDs: []string{ledID.Hex(), controllerID.Hex()},
	}
	kitStale := &domain.Kit{
		ID:         primitive.NewObjectID(),
		Name:       "Kit desactualizado",
		Price:      250, // el led subió de 80 a 100 → debería pasar a 300
		ProductIDs: []string{ledID.Hex(), controllerID.Hex()},
	}
	kitWithDuplicateProduct := &domain.Kit{
		ID:         primitive.NewObjectID(),
		Name:       "Kit con 2 unidades del mismo producto",
		Price:      100,
		ProductIDs: []string{ledID.Hex(), ledID.Hex()}, // 2x led = 200
	}

	kitRepo := &fakeKitRepo{kits: map[primitive.ObjectID]*domain.Kit{
		kitUpToDate.ID:             kitUpToDate,
		kitStale.ID:                kitStale,
		kitWithDuplicateProduct.ID: kitWithDuplicateProduct,
	}}
	productRepo := &fakeProductRepo{products: map[primitive.ObjectID]*domain.Product{
		ledID:        {ID: ledID, BasePrice: 100},
		controllerID: {ID: controllerID, BasePrice: 200},
	}}

	svc := NewKitService(kitRepo, productRepo)
	refreshed, err := svc.RefreshSuggestedPrices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refreshed) != 2 {
		t.Fatalf("expected 2 kits refreshed (stale + duplicate), got %d: %+v", len(refreshed), refreshed)
	}

	if kitUpToDate.Price != 300 {
		t.Errorf("kitUpToDate should NOT have been touched, price = %v", kitUpToDate.Price)
	}
	for _, id := range kitRepo.updated {
		if id == kitUpToDate.ID {
			t.Errorf("kitUpToDate should not have triggered a repo Update call")
		}
	}

	if kitStale.Price != 300 {
		t.Errorf("kitStale.Price = %v, want 300", kitStale.Price)
	}

	if kitWithDuplicateProduct.Price != 200 {
		t.Errorf("kitWithDuplicateProduct.Price = %v, want 200 (2x led)", kitWithDuplicateProduct.Price)
	}
}

func TestRefreshSuggestedPrices_SkipsKitWithMissingProduct(t *testing.T) {
	missingID := primitive.NewObjectID()
	kit := &domain.Kit{
		ID:         primitive.NewObjectID(),
		Name:       "Kit con producto eliminado",
		Price:      500,
		ProductIDs: []string{missingID.Hex()},
	}
	kitRepo := &fakeKitRepo{kits: map[primitive.ObjectID]*domain.Kit{kit.ID: kit}}
	productRepo := &fakeProductRepo{products: map[primitive.ObjectID]*domain.Product{}}

	svc := NewKitService(kitRepo, productRepo)
	refreshed, err := svc.RefreshSuggestedPrices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refreshed) != 0 {
		t.Fatalf("expected kit with missing product to be skipped, got %+v", refreshed)
	}
	if kit.Price != 500 {
		t.Errorf("kit.Price changed unexpectedly: %v", kit.Price)
	}
}
