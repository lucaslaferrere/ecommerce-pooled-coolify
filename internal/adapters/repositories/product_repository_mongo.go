package repositories

import (
	"context"
	"errors"

	"ecommerce-pooled/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductRepositoryMongo es la implementación de ProductRepository con MongoDB
type ProductRepositoryMongo struct {
	collection *mongo.Collection
}

// NewProductRepositoryMongo crea una nueva instancia de ProductRepositoryMongo
func NewProductRepositoryMongo(collection *mongo.Collection) *ProductRepositoryMongo {
	return &ProductRepositoryMongo{
		collection: collection,
	}
}

// Create crea un nuevo producto en la base de datos
func (r *ProductRepositoryMongo) Create(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return errors.New("producto no puede ser nil")
	}

	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return err
	}

	// Asignar el ID generado por MongoDB
	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID obtiene un producto por su ID
func (r *ProductRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	if id.IsZero() {
		return nil, errors.New("ID no puede ser vacío")
	}

	var product domain.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

// Update actualiza un producto existente
func (r *ProductRepositoryMongo) Update(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return errors.New("producto no puede ser nil")
	}

	if product.ID.IsZero() {
		return errors.New("ID del producto no puede ser vacío")
	}

	filter := bson.M{"_id": product.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"description": product.Description,
			"base_price":  product.BasePrice,
			"stock":       product.Stock,
			"category":    product.Category,
			"brand":       product.Brand,
			"images":      product.Images,
			"variants":    product.Variants,
			"specs":       product.Specs,
			"updated_at":  product.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("producto no encontrado")
	}

	return nil
}

// Delete elimina un producto por su ID (hard delete)
func (r *ProductRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	if id.IsZero() {
		return errors.New("ID no puede ser vacío")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("producto no encontrado")
	}

	return nil
}

// List obtiene una lista de productos con filtros opcionales y paginación
func (r *ProductRepositoryMongo) List(ctx context.Context, filter map[string]interface{}, skip int64, limit int64) ([]*domain.Product, error) {
	// Construir filtro BSON desde el mapa
	bsonFilter := bson.M{}
	for k, v := range filter {
		bsonFilter[k] = v
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// GetByCategory obtiene productos por categoría
func (r *ProductRepositoryMongo) GetByCategory(ctx context.Context, category string, skip int64, limit int64) ([]*domain.Product, error) {
	if category == "" {
		return nil, errors.New("categoría no puede ser vacía")
	}

	filter := bson.M{"category": category}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// GetByBrand obtiene productos por marca
func (r *ProductRepositoryMongo) GetByBrand(ctx context.Context, brand string, skip int64, limit int64) ([]*domain.Product, error) {
	if brand == "" {
		return nil, errors.New("marca no puede ser vacía")
	}

	filter := bson.M{"brand": brand}
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// DecrementVariantStock reduce el stock atómicamente usando un filtro condicional ($elemMatch + $inc).
// Retorna error si el stock es insuficiente o la variante/producto no existe.
func (r *ProductRepositoryMongo) DecrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error {
	if productID.IsZero() {
		return errors.New("ID de producto no puede ser vacío")
	}
	if sku == "" {
		return errors.New("SKU no puede ser vacío")
	}
	if quantity <= 0 {
		return errors.New("cantidad debe ser mayor a 0")
	}

	filter := bson.M{
		"_id": productID,
		"variants": bson.M{
			"$elemMatch": bson.M{
				"sku":   sku,
				"stock": bson.M{"$gte": quantity},
			},
		},
	}
	update := bson.M{"$inc": bson.M{"variants.$.stock": -quantity}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("stock insuficiente o variante no encontrada")
	}
	return nil
}

// IncrementVariantStock incrementa el stock de una variante (utilizado para rollback de checkout).
func (r *ProductRepositoryMongo) IncrementVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, quantity int) error {
	if productID.IsZero() {
		return errors.New("ID de producto no puede ser vacío")
	}
	if sku == "" {
		return errors.New("SKU no puede ser vacío")
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"v.sku": sku}},
	})
	filter := bson.M{"_id": productID}
	update := bson.M{"$inc": bson.M{"variants.$[v].stock": quantity}}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpdateVariantStock actualiza el stock de una variante específica dentro del array Variants
// Utiliza arrayFilters de MongoDB para actualizar elementos dentro de arrays anidados
func (r *ProductRepositoryMongo) UpdateVariantStock(ctx context.Context, productID primitive.ObjectID, sku string, newStock int) error {
	if productID.IsZero() {
		return errors.New("ID de producto no puede ser vacío")
	}

	if sku == "" {
		return errors.New("SKU no puede ser vacío")
	}

	// Usar arrayFilters para actualizar solo la variante con el SKU especificado
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"variant.sku": sku},
		},
	})

	filter := bson.M{"_id": productID}
	update := bson.M{
		"$set": bson.M{
			"variants.$[variant].stock": newStock,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("producto no encontrado")
	}

	if result.ModifiedCount == 0 {
		return errors.New("variante con SKU no encontrada o stock no cambió")
	}

	return nil
}

// DecrementProductStock reduce el stock a nivel producto de forma atómica.
// Usado para productos sin variantes.
func (r *ProductRepositoryMongo) DecrementProductStock(ctx context.Context, productID primitive.ObjectID, quantity int) error {
	if productID.IsZero() {
		return errors.New("ID de producto no puede ser vacío")
	}
	filter := bson.M{
		"_id":   productID,
		"stock": bson.M{"$gte": quantity},
	}
	result, err := r.collection.UpdateOne(ctx, filter, bson.M{"$inc": bson.M{"stock": -quantity}})
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("stock insuficiente")
	}
	return nil
}

// IncrementProductStock incrementa el stock a nivel producto (rollback de checkout).
func (r *ProductRepositoryMongo) IncrementProductStock(ctx context.Context, productID primitive.ObjectID, quantity int) error {
	if productID.IsZero() {
		return errors.New("ID de producto no puede ser vacío")
	}
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": productID},
		bson.M{"$inc": bson.M{"stock": quantity}},
	)
	return err
}
