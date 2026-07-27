package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShippingDetails contiene la información de entrega de un pedido
type ShippingDetails struct {
	Address    string `bson:"address" json:"address"`
	City       string `bson:"city" json:"city"`
	Province   string `bson:"province,omitempty" json:"province,omitempty"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
}

// FacturaA contiene los datos para emisión de Factura A
type FacturaA struct {
	RazonSocial string `bson:"razon_social" json:"razon_social"`
	CUIT        string `bson:"cuit" json:"cuit"`
}

// SalesBucket representa ventas pagadas agregadas por período (mes o día).
// Period es "2006-01" con granularidad mes, o "2006-01-02" con granularidad día.
type SalesBucket struct {
	Period  string  `bson:"_id" json:"period"`
	Revenue float64 `bson:"revenue" json:"revenue"`
	Orders  int64   `bson:"orders" json:"orders"`
}

// SalesTotals son los totales de un período (para las KPIs del dashboard).
type SalesTotals struct {
	Revenue    float64 `json:"revenue"`
	Orders     int64   `json:"orders"`      // todas las órdenes del período
	PaidOrders int64   `json:"paid_orders"` // solo estados pagados
	AOV        float64 `json:"aov"`         // revenue / paid_orders
}

// CategoryRevenue: facturación por categoría de producto.
type CategoryRevenue struct {
	Category string  `bson:"_id" json:"category"`
	Revenue  float64 `bson:"revenue" json:"revenue"`
}

// ProductSales: unidades y facturación por producto.
type ProductSales struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Units     int64   `json:"units"`
	Revenue   float64 `json:"revenue"`
}

// SalesOverview es la respuesta de GET /admin/analytics/overview.
type SalesOverview struct {
	Current     SalesTotals       `json:"current"`
	Previous    SalesTotals       `json:"previous"`
	Series      []SalesBucket     `json:"series"`
	ByCategory  []CategoryRevenue `json:"by_category"`
	TopProducts []ProductSales    `json:"top_products"`
}

// BuyerExportRow es una fila del export de compradores (una por persona).
type BuyerExportRow struct {
	Name         string
	Email        string
	Phone        string
	Province     string
	Products     string // resumen "Producto x3, Otro x1"
	TotalSpent   float64
	LastPurchase time.Time
	OrdersCount  int
}

// Order representa la entidad de orden/pedido en la aplicación
type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	Items           []CartItem         `bson:"items" json:"items"`
	Total           float64            `bson:"total" json:"total"`
	Status          string             `bson:"status" json:"status"` // pending | paid | processing | shipped | delivered | cancelled
	CustomerName    string             `bson:"customer_name" json:"customer_name"`
	CustomerEmail   string             `bson:"customer_email" json:"customer_email"`
	CustomerPhone   string             `bson:"customer_phone" json:"customer_phone"`
	DniCuit         string             `bson:"dni_cuit,omitempty" json:"dni_cuit,omitempty"`
	Notes           string             `bson:"notes,omitempty" json:"notes,omitempty"`
	PaymentMethod   string             `bson:"payment_method,omitempty" json:"payment_method,omitempty"`
	DeliveryMethod  string             `bson:"delivery_method,omitempty" json:"delivery_method,omitempty"`
	ShippingDetails ShippingDetails    `bson:"shipping_details" json:"shipping_details"`
	FacturaA        *FacturaA          `bson:"factura_a,omitempty" json:"factura_a,omitempty"`
	ShippingCost    float64            `bson:"shipping_cost,omitempty" json:"shipping_cost,omitempty"`
	Discount        float64            `bson:"discount,omitempty" json:"discount,omitempty"`
	CouponCode      string             `bson:"coupon_code,omitempty"     json:"coupon_code,omitempty"`
	CouponDiscount  float64            `bson:"coupon_discount,omitempty" json:"coupon_discount,omitempty"`
	TrackingNumber  string             `bson:"tracking_number,omitempty" json:"tracking_number,omitempty"`
	InvoiceSentAt   *time.Time         `bson:"invoice_sent_at,omitempty" json:"invoice_sent_at,omitempty"`
	PreferenceID    string             `bson:"preference_id,omitempty" json:"preference_id,omitempty"`
	PaymentID       string             `bson:"payment_id,omitempty" json:"payment_id,omitempty"`
	SessionID       string             `bson:"session_id,omitempty" json:"session_id,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}
