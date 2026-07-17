package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event representa un evento de comportamiento del usuario (analytics).
type Event struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Type      string                 `bson:"type" json:"type"`
	SessionID string                 `bson:"session_id" json:"session_id"`
	Payload   map[string]interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	IP        string                 `bson:"ip,omitempty" json:"ip,omitempty"`
	Country   string                 `bson:"country,omitempty" json:"country,omitempty"`
	Province  string                 `bson:"province,omitempty" json:"province,omitempty"`
	City      string                 `bson:"city,omitempty" json:"city,omitempty"`
	Lat       float64                `bson:"lat,omitempty" json:"lat,omitempty"`
	Lng       float64                `bson:"lng,omitempty" json:"lng,omitempty"`
	CreatedAt time.Time              `bson:"created_at" json:"created_at"`
}

// DayCount representa el conteo de eventos en un día.
type DayCount struct {
	Date  string `json:"date" bson:"_id"`
	Count int64  `json:"count" bson:"count"`
}

// ItemStat representa el conteo agrupado por un campo de payload.
type ItemStat struct {
	ID    string `json:"id" bson:"_id"`
	Name  string `json:"name" bson:"name"`
	Count int64  `json:"count" bson:"count"`
}

// ProductStat combina vistas y cart_adds de un producto.
type ProductStat struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Views    int64  `json:"views"`
	CartAdds int64  `json:"cart_adds"`
}

// AnalyticsReport es la respuesta del endpoint GET /admin/analytics.
type AnalyticsReport struct {
	Period         string        `json:"period"`
	CartAdds       int64         `json:"cart_adds"`
	CheckoutStarts int64         `json:"checkout_starts"`
	ProductViews   int64         `json:"product_views"`
	PaidOrders     int64         `json:"paid_orders"`
	ConversionRate float64       `json:"conversion_rate"`
	ByDay          []DayCount    `json:"by_day"`
	TopProducts    []ProductStat `json:"top_products"`
}

// PageStat representa el conteo de visitas a una URL.
type PageStat struct {
	URL   string `json:"url"`
	Count int64  `json:"count"`
}

// LocationStat: sesiones únicas por país/provincia/ciudad, con coordenadas para el mapa.
type LocationStat struct {
	Country  string  `bson:"country" json:"country"`
	Province string  `bson:"province" json:"province"`
	City     string  `bson:"city" json:"city"`
	Lat      float64 `bson:"lat" json:"lat"`
	Lng      float64 `bson:"lng" json:"lng"`
	Sessions int64   `bson:"sessions" json:"sessions"`
}

// NewVsReturning: órdenes de clientes nuevos vs recurrentes.
type NewVsReturning struct {
	Nuevos      int64 `json:"nuevos"`
	Recurrentes int64 `json:"recurrentes"`
}

// RealtimeSnapshot: respuesta de GET /admin/realtime/snapshot.
type RealtimeSnapshot struct {
	VisitantesActivos    int            `json:"visitantes_activos"`
	VisitantesUltimaHora int64          `json:"visitantes_ultima_hora"`
	CarritosActivos      int64          `json:"carritos_activos"`
	EnPago               int64          `json:"en_pago"`
	ComprasHoy           int64          `json:"compras_hoy"`
	SesionesPorUbicacion []LocationStat `json:"sesiones_por_ubicacion"`
	NuevosVsRecurrentes  NewVsReturning `json:"nuevos_vs_recurrentes"`
}

// TrafficReport es la respuesta del endpoint GET /admin/traffic.
type TrafficReport struct {
	ViewsToday     int64      `json:"views_today"`
	ViewsWeek      int64      `json:"views_week"`
	ViewsMonth     int64      `json:"views_month"`
	UniqueSessions int64      `json:"unique_sessions"`
	TopPages       []PageStat `json:"top_pages"`
}
