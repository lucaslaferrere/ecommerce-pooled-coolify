package repositories

import (
	"context"
	"errors"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderRepositoryMongo es la implementación de OrderRepository con MongoDB
type OrderRepositoryMongo struct {
	collection *mongo.Collection
}

// NewOrderRepositoryMongo crea una nueva instancia de OrderRepositoryMongo
func NewOrderRepositoryMongo(collection *mongo.Collection) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{collection: collection}
}

// paidOrderStatuses son los estados que cuentan como una orden concretada.
// Una orden pending o cancelled no consume usos de cupón.
var paidOrderStatuses = bson.A{"paid", "processing", "shipped", "delivered"}

// CountPaidByCoupon cuenta cuántas órdenes pagadas usaron un cupón (límite total).
func (r *OrderRepositoryMongo) CountPaidByCoupon(ctx context.Context, code string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"coupon_code": code,
		"status":      bson.M{"$in": paidOrderStatuses},
	})
}

// CountPaidByCouponAndUser cuenta cuántas órdenes pagadas de un usuario usaron un
// cupón (límite por cliente).
func (r *OrderRepositoryMongo) CountPaidByCouponAndUser(ctx context.Context, code string, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"coupon_code": code,
		"user_id":     userID,
		"status":      bson.M{"$in": paidOrderStatuses},
	})
}

// CountPaidGroupedByCoupon devuelve un mapa coupon_code → cantidad de órdenes
// pagadas, en una sola agregación. Se usa para mostrar el contador en el admin.
func (r *OrderRepositoryMongo) CountPaidGroupedByCoupon(ctx context.Context) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"coupon_code": bson.M{"$nin": bson.A{nil, ""}},
			"status":      bson.M{"$in": paidOrderStatuses},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$coupon_code",
			"count": bson.M{"$sum": 1},
		}}},
	}
	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make(map[string]int64)
	for cur.Next(ctx) {
		var row struct {
			Code  string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.Code] = row.Count
	}
	return out, cur.Err()
}

// Create crea una nueva orden en la base de datos
func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return errors.New("orden no puede ser nil")
	}
	result, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	order.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID obtiene una orden por su ID
func (r *OrderRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	if id.IsZero() {
		return nil, errors.New("ID no puede ser vacío")
	}
	var order domain.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByUserID obtiene las órdenes de un usuario con paginación, ordenadas por fecha desc
func (r *OrderRepositoryMongo) GetByUserID(ctx context.Context, userID primitive.ObjectID, skip int64, limit int64) ([]*domain.Order, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// Update actualiza una orden existente
func (r *OrderRepositoryMongo) Update(ctx context.Context, order *domain.Order) error {
	if order == nil || order.ID.IsZero() {
		return errors.New("orden inválida")
	}
	order.UpdatedAt = time.Now()
	filter := bson.M{"_id": order.ID}
	update := bson.M{"$set": bson.M{
		"items":           order.Items,
		"total":           order.Total,
		"status":          order.Status,
		"preference_id":   order.PreferenceID,
		"payment_id":      order.PaymentID,
		"tracking_number": order.TrackingNumber,
		"updated_at":      order.UpdatedAt,
	}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("orden no encontrada")
	}
	return nil
}

// UpdateStatus actualiza únicamente el estado de una orden
func (r *OrderRepositoryMongo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	if id.IsZero() {
		return errors.New("ID no puede ser vacío")
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("orden no encontrada")
	}
	return nil
}

// Delete elimina una orden por su ID
func (r *OrderRepositoryMongo) Delete(ctx context.Context, id primitive.ObjectID) error {
	if id.IsZero() {
		return errors.New("ID no puede ser vacío")
	}
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("orden no encontrada")
	}
	return nil
}

// List obtiene todas las órdenes con paginación, ordenadas por fecha desc
func (r *OrderRepositoryMongo) List(ctx context.Context, skip int64, limit int64) ([]*domain.Order, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// CountByStatusInPeriod cuenta órdenes con alguno de los estados dados entre from y to.
func (r *OrderRepositoryMongo) CountByStatusInPeriod(ctx context.Context, statuses []string, from, to time.Time) (int64, error) {
	filter := bson.M{
		"status":     bson.M{"$in": statuses},
		"created_at": bson.M{"$gte": from, "$lte": to},
	}
	return r.collection.CountDocuments(ctx, filter)
}

// FindPendingOlderThan devuelve órdenes en estado "pending" creadas antes de `before`.
func (r *OrderRepositoryMongo) FindPendingOlderThan(ctx context.Context, before time.Time) ([]*domain.Order, error) {
	filter := bson.M{
		"status":     "pending",
		"created_at": bson.M{"$lt": before},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// Count devuelve el total de órdenes, opcionalmente filtrado por status.
func (r *OrderRepositoryMongo) Count(ctx context.Context, status string) (int64, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	return r.collection.CountDocuments(ctx, filter)
}

// paidStatusList son los estados que cuentan como venta concretada.
var paidStatusList = bson.A{"paid", "processing", "shipped", "delivered"}

// FindByStatuses devuelve todas las órdenes con alguno de los estados, ordenadas
// por fecha ascendente. Sin paginar: pensado para exports puntuales del admin.
func (r *OrderRepositoryMongo) FindByStatuses(ctx context.Context, statuses []string) ([]*domain.Order, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"status": bson.M{"$in": statuses}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []*domain.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// SalesTotals devuelve totales del período en una sola pasada: todas las órdenes,
// las pagadas y su facturación.
func (r *OrderRepositoryMongo) SalesTotals(ctx context.Context, from, to time.Time) (domain.SalesTotals, error) {
	isPaid := bson.M{"$in": bson.A{"$status", paidStatusList}}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": from, "$lte": to}}}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"orders":      bson.M{"$sum": 1},
			"paid_orders": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 1, 0}}},
			"revenue":     bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, "$total", 0}}},
		}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return domain.SalesTotals{}, err
	}
	defer cursor.Close(ctx)
	var res []struct {
		Orders     int64   `bson:"orders"`
		PaidOrders int64   `bson:"paid_orders"`
		Revenue    float64 `bson:"revenue"`
	}
	if err = cursor.All(ctx, &res); err != nil || len(res) == 0 {
		return domain.SalesTotals{}, err
	}
	t := domain.SalesTotals{Revenue: res[0].Revenue, Orders: res[0].Orders, PaidOrders: res[0].PaidOrders}
	if t.PaidOrders > 0 {
		t.AOV = t.Revenue / float64(t.PaidOrders)
	}
	return t, nil
}

// SalesByCategory agrega facturación pagada por categoría (join con products).
// Los ítems que no matchean un producto (ej. kits) caen en "otros".
func (r *OrderRepositoryMongo) SalesByCategory(ctx context.Context, from, to time.Time) ([]domain.CategoryRevenue, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status":     bson.M{"$in": paidStatusList},
			"created_at": bson.M{"$gte": from, "$lte": to},
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "items.product_id",
			"foreignField": "_id",
			"as":           "prod",
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$ifNull": bson.A{
				bson.M{"$arrayElemAt": bson.A{"$prod.category", 0}}, "otros",
			}},
			"revenue": bson.M{"$sum": bson.M{"$multiply": bson.A{"$items.unit_price", "$items.quantity"}}},
		}}},
		{{Key: "$sort", Value: bson.M{"revenue": -1}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	out := []domain.CategoryRevenue{}
	return out, cursor.All(ctx, &out)
}

// TopSellingProducts agrupa los ítems vendidos por producto, ordenados por unidades.
func (r *OrderRepositoryMongo) TopSellingProducts(ctx context.Context, from, to time.Time, limit int) ([]domain.ProductSales, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status":     bson.M{"$in": paidStatusList},
			"created_at": bson.M{"$gte": from, "$lte": to},
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":     "$items.product_id",
			"name":    bson.M{"$first": "$items.name"},
			"units":   bson.M{"$sum": "$items.quantity"},
			"revenue": bson.M{"$sum": bson.M{"$multiply": bson.A{"$items.unit_price", "$items.quantity"}}},
		}}},
		{{Key: "$sort", Value: bson.M{"units": -1}}},
		{{Key: "$limit", Value: int64(limit)}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var raw []struct {
		ID      primitive.ObjectID `bson:"_id"`
		Name    string             `bson:"name"`
		Units   int64              `bson:"units"`
		Revenue float64            `bson:"revenue"`
	}
	if err = cursor.All(ctx, &raw); err != nil {
		return nil, err
	}
	out := make([]domain.ProductSales, 0, len(raw))
	for _, p := range raw {
		out = append(out, domain.ProductSales{
			ProductID: p.ID.Hex(), Name: p.Name, Units: p.Units, Revenue: p.Revenue,
		})
	}
	return out, nil
}

// SalesByPeriod agrega ventas pagadas por mes ("month") o día ("day") en el rango.
// Los buckets se calculan en timezone de Argentina para que el corte de día/mes
// coincida con lo que ve el admin.
func (r *OrderRepositoryMongo) SalesByPeriod(ctx context.Context, from, to time.Time, granularity string) ([]domain.SalesBucket, error) {
	format := "%Y-%m"
	if granularity == "day" {
		format = "%Y-%m-%d"
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status":     bson.M{"$in": bson.A{"paid", "processing", "shipped", "delivered"}},
			"created_at": bson.M{"$gte": from, "$lte": to},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$dateToString": bson.M{
				"format":   format,
				"date":     "$created_at",
				"timezone": "America/Argentina/Buenos_Aires",
			}},
			"revenue": bson.M{"$sum": "$total"},
			"orders":  bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	out := []domain.SalesBucket{} // slice vacío (no nil) → JSON [] en vez de null
	return out, cursor.All(ctx, &out)
}

// CountNewVsReturning agrupa órdenes pagadas por user_id: (nuevos, total), donde
// nuevos = clientes distintos (1ra orden c/u) y total = todas las órdenes pagadas.
func (r *OrderRepositoryMongo) CountNewVsReturning(ctx context.Context) (nuevos, total int64, err error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status": bson.M{"$in": bson.A{"paid", "processing", "shipped", "delivered"}},
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$user_id", "c": bson.M{"$sum": 1}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "nuevos": bson.M{"$sum": 1}, "total": bson.M{"$sum": "$c"}}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)
	var res []struct {
		Nuevos int64 `bson:"nuevos"`
		Total  int64 `bson:"total"`
	}
	if err = cursor.All(ctx, &res); err != nil || len(res) == 0 {
		return 0, 0, err
	}
	return res[0].Nuevos, res[0].Total, nil
}

// GetByStatus obtiene órdenes filtradas por estado con paginación
func (r *OrderRepositoryMongo) GetByStatus(ctx context.Context, status string, skip int64, limit int64) ([]*domain.Order, error) {
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
