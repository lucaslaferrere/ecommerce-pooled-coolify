# Abandoned Cart Recovery — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Permitir que el admin vea los carritos de clientes logueados que no compraron y les envíe un email de recuperación con un cupón elegido de la lista, usando un template editable.

**Architecture:** El carrito del cliente logueado se persiste en una colección `carts` (Mongo) sincronizada desde el frontend en cada cambio. Un panel admin lista esos carritos, permite verlos y disparar un email (Resend) armado con un template global editable + variables del cupón. Se reutiliza el sistema de cupones y el `EmailService` existentes.

**Tech Stack:** Go + Gin + MongoDB (backend `ecommerce-pooled`), React + TypeScript + Vite + Zustand + React Query (frontend `poolside-commerce-hub`), Resend para email.

## Global Constraints

- Backend production deploy: `git push coolify v2` (repo `ecommerce-pooled-coolify`). Push también a `origin v2`.
- Frontend deploy: `git push origin v2` (`poolside-commerce-hub`).
- Nunca pushear archivos `.env`.
- Verificación por tarea: backend `go build ./...`; frontend `npx tsc --noEmit`; + prueba manual descrita. No hay framework de tests unitarios en uso.
- Commits: conventional commits, sin atribución AI/Co-Authored-By.
- Solo se persisten y recuperan carritos de usuarios **logueados**.
- Idioma de artefactos (código, UI, comentarios): español neutro donde el proyecto ya lo usa (UI en español, comentarios en español), identificadores en inglés.

---

## File Structure

**Backend (`ecommerce-pooled`):**
- Create: `internal/core/domain/cart.go` — entidad `Cart` (carrito persistido).
- Create: `internal/adapters/repositories/cart_repository_mongo.go` — CRUD de `carts`.
- Create: `internal/core/services/cart_storage_service.go` — lógica de persistencia del carrito.
- Create: `internal/adapters/handlers/cart_storage_handler.go` — endpoints cliente (`PUT/DELETE /cart`) y admin (`GET /admin/carts`, `POST /admin/carts/:userID/send-coupon`).
- Create: `internal/core/domain/setting.go` — entidad `Setting` (key/value para el template).
- Create: `internal/adapters/repositories/setting_repository_mongo.go` — get/upsert de settings.
- Create: `internal/core/services/setting_service.go` — settings + default del template.
- Create: `internal/adapters/handlers/setting_handler.go` — endpoints admin del template.
- Modify: `internal/core/services/email_service.go` — `SendAbandonedCartEmail`.
- Modify: `internal/adapters/handlers/order_handler.go` — vaciar carrito server-side en checkout.
- Modify: `internal/app/server.go` — wiring y rutas.

**Frontend (`poolside-commerce-hub`):**
- Create: `src/lib/cartSync.ts` — helper de sincronización del carrito con el backend.
- Modify: `src/store/cart.ts` — disparar sync en cambios.
- Modify: `src/context/AuthContext.tsx` — subir carrito al login.
- Create: `src/pages/admin/CartsPage.tsx` — sección Carritos.
- Create: `src/components/admin/CartDetailDialog.tsx` — ver productos del carrito.
- Create: `src/components/admin/SendCouponDialog.tsx` — elegir cupón + preview + enviar.
- Create: `src/components/admin/AbandonedEmailTemplateDialog.tsx` — editor del template.
- Modify: `src/App.tsx`, `src/pages/admin/AdminLayout.tsx` — ruta + link sidebar.

---

## Task 1: Entidad Cart + repositorio Mongo

**Files:**
- Create: `internal/core/domain/cart.go`
- Create: `internal/adapters/repositories/cart_repository_mongo.go`

**Interfaces:**
- Produces: `domain.Cart{ UserID, Items []CartItem, UpdatedAt, LastReminderSentAt *time.Time }`
- Produces: `CartRepositoryMongo` con `Upsert(ctx, userID, items) error`, `DeleteByUser(ctx, userID) error`, `GetByUser(ctx, userID) (*Cart, error)`, `ListNonEmpty(ctx) ([]Cart, error)`, `SetReminderSentAt(ctx, userID, t) error`.

- [ ] **Step 1: Crear la entidad `Cart`**

`internal/core/domain/cart.go`:
```go
package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Cart es el carrito persistido de un usuario logueado. Se sincroniza desde el
// frontend en cada cambio y se vacía cuando el usuario concreta una compra.
type Cart struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"                json:"id,omitempty"`
	UserID             primitive.ObjectID `bson:"user_id"                      json:"user_id"`
	Items              []CartItem         `bson:"items"                        json:"items"`
	UpdatedAt          time.Time          `bson:"updated_at"                   json:"updated_at"`
	LastReminderSentAt *time.Time         `bson:"last_reminder_sent_at,omitempty" json:"last_reminder_sent_at,omitempty"`
}
```

- [ ] **Step 2: Crear el repositorio Mongo**

`internal/adapters/repositories/cart_repository_mongo.go`:
```go
package repositories

import (
	"context"
	"time"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CartRepositoryMongo struct {
	collection *mongo.Collection
}

func NewCartRepositoryMongo(col *mongo.Collection) *CartRepositoryMongo {
	return &CartRepositoryMongo{collection: col}
}

// Upsert reemplaza los items del carrito del usuario (crea el doc si no existe).
func (r *CartRepositoryMongo) Upsert(ctx context.Context, userID primitive.ObjectID, items []domain.CartItem) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"items": items, "updated_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *CartRepositoryMongo) DeleteByUser(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

func (r *CartRepositoryMongo) GetByUser(ctx context.Context, userID primitive.ObjectID) (*domain.Cart, error) {
	var c domain.Cart
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

// ListNonEmpty devuelve los carritos que tienen al menos un item.
func (r *CartRepositoryMongo) ListNonEmpty(ctx context.Context) ([]domain.Cart, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cur, err := r.collection.Find(ctx, bson.M{"items.0": bson.M{"$exists": true}}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Cart
	return out, cur.All(ctx, &out)
}

func (r *CartRepositoryMongo) SetReminderSentAt(ctx context.Context, userID primitive.ObjectID, t time.Time) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"last_reminder_sent_at": t}},
	)
	return err
}
```

- [ ] **Step 3: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: sin salida (compila).

- [ ] **Step 4: Commit**

```bash
git add internal/core/domain/cart.go internal/adapters/repositories/cart_repository_mongo.go
git commit -m "feat: entidad Cart y repositorio Mongo para carritos persistidos"
```

---

## Task 2: Settings (colección) + template por default

**Files:**
- Create: `internal/core/domain/setting.go`
- Create: `internal/adapters/repositories/setting_repository_mongo.go`
- Create: `internal/core/services/setting_service.go`
- Create: `internal/adapters/handlers/setting_handler.go`
- Modify: `internal/app/server.go`

**Interfaces:**
- Produces: `SettingService.GetAbandonedCartBody(ctx) (string, error)` (default si vacío), `.SetAbandonedCartBody(ctx, body) error`, `.DefaultAbandonedCartBody() string`.
- Produces: rutas admin `GET/PUT /admin/settings/abandoned-cart-email`.

Esta task es independiente y deja el server compilando. Va antes del handler de
carrito porque ese handler consume `SettingService`.

- [ ] **Step 1: Entidad Setting**

`internal/core/domain/setting.go`:
```go
package domain

// Setting es un par key/value para configuración global editable.
type Setting struct {
	Key   string `bson:"key"   json:"key"`
	Value string `bson:"value" json:"value"`
}
```

- [ ] **Step 2: Repositorio de settings**

`internal/adapters/repositories/setting_repository_mongo.go`:
```go
package repositories

import (
	"context"

	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingRepositoryMongo struct {
	collection *mongo.Collection
}

func NewSettingRepositoryMongo(col *mongo.Collection) *SettingRepositoryMongo {
	return &SettingRepositoryMongo{collection: col}
}

// Get devuelve el value de una key, o "" si no existe.
func (r *SettingRepositoryMongo) Get(ctx context.Context, key string) (string, error) {
	var s domain.Setting
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	return s.Value, err
}

func (r *SettingRepositoryMongo) Set(ctx context.Context, key, value string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"key": key},
		bson.M{"$set": bson.M{"value": value}},
		options.Update().SetUpsert(true),
	)
	return err
}
```

- [ ] **Step 3: Servicio de settings con el default del template**

`internal/core/services/setting_service.go`:
```go
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
```

- [ ] **Step 4: Handler de settings**

`internal/adapters/handlers/setting_handler.go`:
```go
package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	svc *services.SettingService
}

func NewSettingHandler(svc *services.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

// GET /admin/settings/abandoned-cart-email
func (h *SettingHandler) GetAbandonedCartEmail(c *gin.Context) {
	body, err := h.svc.GetAbandonedCartBody(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"body":         body,
		"default_body": h.svc.DefaultAbandonedCartBody(),
	})
}

// PUT /admin/settings/abandoned-cart-email
func (h *SettingHandler) SetAbandonedCartEmail(c *gin.Context) {
	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetAbandonedCartBody(c.Request.Context(), req.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plantilla guardada"})
}
```

- [ ] **Step 5: Wiring en server.go**

Repos:
```go
	settingRepo    := repositories.NewSettingRepositoryMongo(db.Collection("settings"))
```
Servicios:
```go
	settingService := services.NewSettingService(settingRepo)
```
Handlers:
```go
	settingHandler := handlers.NewSettingHandler(settingService)
```
Rutas admin (dentro del grupo `admin`):
```go
		admin.GET("/settings/abandoned-cart-email", settingHandler.GetAbandonedCartEmail)
		admin.PUT("/settings/abandoned-cart-email", settingHandler.SetAbandonedCartEmail)
```

- [ ] **Step 6: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 7: Commit**

```bash
git add internal/core/domain/setting.go internal/adapters/repositories/setting_repository_mongo.go internal/core/services/setting_service.go internal/adapters/handlers/setting_handler.go internal/app/server.go
git commit -m "feat: settings editables con template de email de recuperacion"
```

---

## Task 3: Email de recuperación + GetByID de cupón

**Files:**
- Modify: `internal/core/services/email_service.go`
- Modify: `internal/core/services/coupon_service.go`
- Modify: `internal/adapters/repositories/coupon_repository_mongo.go`

**Interfaces:**
- Produces: `EmailService.SendAbandonedCartEmail(toEmail, bodyText string, coupon *domain.Coupon, items []domain.CartItem, total float64) error`.
- Produces: `CouponService.GetByID(ctx, id) (*domain.Coupon, error)`.

Métodos exportados nuevos: compilan aunque todavía no se usen. Los consume el
handler de la Task 4.

- [ ] **Step 1: Método de email en email_service.go**

Agregar en `internal/core/services/email_service.go` (usa `s.Send(to, subject, html)` y `s.Enabled()`; imports `fmt`, `strings` ya presentes; `domain` ya importado):
```go
// SendAbandonedCartEmail envía el email de recuperación de carrito con el cupón.
// bodyText es el template editable; se interpolan las variables y se envuelve en HTML.
func (s *EmailService) SendAbandonedCartEmail(toEmail, bodyText string, coupon *domain.Coupon, items []domain.CartItem, total float64) error {
	if !s.Enabled() {
		return nil
	}

	// Construir el listado de productos en texto.
	var productsText strings.Builder
	for _, it := range items {
		name := it.Name
		if name == "" {
			name = it.VariantSKU
		}
		fmt.Fprintf(&productsText, "- %s x%d\n", name, it.Quantity)
	}

	vencimiento := "sin vencimiento"
	if coupon.ExpiresAt != nil {
		vencimiento = coupon.ExpiresAt.Format("02/01/2006")
	}
	descuento := fmt.Sprintf("%g%%", coupon.DiscountPercent)
	totalStr := fmt.Sprintf("$%.0f", total)

	replacer := strings.NewReplacer(
		"{{codigo}}", coupon.Code,
		"{{descuento}}", descuento,
		"{{vencimiento}}", vencimiento,
		"{{productos}}", productsText.String(),
		"{{total}}", totalStr,
	)
	filled := replacer.Replace(bodyText)

	// Convertir saltos de línea a <br> y envolver en HTML simple.
	htmlBody := strings.ReplaceAll(filled, "\n", "<br>")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="es"><head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:-apple-system,Segoe UI,Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
        <tr><td style="background:#0ea5e9;padding:24px 32px;">
          <h1 style="margin:0;color:#fff;font-size:18px;">🛒 Tu carrito te espera</h1>
        </td></tr>
        <tr><td style="padding:28px 32px;color:#374151;font-size:14px;line-height:1.6;">
          %s
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`, htmlBody)

	subject := fmt.Sprintf("Te dejamos un %s de descuento 🎁", descuento)
	return s.Send(toEmail, subject, html)
}
```

- [ ] **Step 2: `GetByID` en CouponService y repo**

En `internal/core/services/coupon_service.go`:
```go
func (s *CouponService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Coupon, error) {
	return s.repo.GetByID(ctx, id)
}
```
En `internal/adapters/repositories/coupon_repository_mongo.go`, **solo si no existe ya** un `GetByID`:
```go
func (r *CouponRepositoryMongo) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Coupon, error) {
	var c domain.Coupon
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}
```

- [ ] **Step 3: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 4: Commit**

```bash
git add internal/core/services/email_service.go internal/core/services/coupon_service.go internal/adapters/repositories/coupon_repository_mongo.go
git commit -m "feat: email de recuperacion de carrito y GetByID de cupon"
```

---

## Task 4: Servicio + handler de carrito (cliente y admin) + rutas

**Files:**
- Create: `internal/core/services/cart_storage_service.go`
- Create: `internal/adapters/handlers/cart_storage_handler.go`
- Modify: `internal/app/server.go`

**Interfaces:**
- Consumes: `CartRepositoryMongo` (Task 1), `CouponService.GetByID` + `EmailService.SendAbandonedCartEmail` (Task 3), `SettingService` (Task 2), `UserService.GetUser` (existente).
- Produces: `CartStorageService` con `Save/Clear/List/Get/MarkReminderSent`.
- Produces: rutas `PUT /cart`, `DELETE /cart`, `GET /admin/carts`, `POST /admin/carts/:userID/send-coupon`.

- [ ] **Step 1: Crear el servicio**

`internal/core/services/cart_storage_service.go`:
```go
package services

import (
	"context"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartStorageService struct {
	repo *repositories.CartRepositoryMongo
}

func NewCartStorageService(repo *repositories.CartRepositoryMongo) *CartStorageService {
	return &CartStorageService{repo: repo}
}

func (s *CartStorageService) Save(ctx context.Context, userID primitive.ObjectID, items []domain.CartItem) error {
	return s.repo.Upsert(ctx, userID, items)
}

func (s *CartStorageService) Clear(ctx context.Context, userID primitive.ObjectID) error {
	return s.repo.DeleteByUser(ctx, userID)
}

func (s *CartStorageService) List(ctx context.Context) ([]domain.Cart, error) {
	return s.repo.ListNonEmpty(ctx)
}

func (s *CartStorageService) Get(ctx context.Context, userID primitive.ObjectID) (*domain.Cart, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *CartStorageService) MarkReminderSent(ctx context.Context, userID primitive.ObjectID) error {
	return s.repo.SetReminderSentAt(ctx, userID, time.Now())
}
```

- [ ] **Step 2: Crear el handler (cliente + admin)**

`internal/adapters/handlers/cart_storage_handler.go`:
```go
package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartStorageHandler struct {
	svc          *services.CartStorageService
	couponSvc    *services.CouponService
	emailService *services.EmailService
	userService  *services.UserService
	settingSvc   *services.SettingService
}

func NewCartStorageHandler(
	svc *services.CartStorageService,
	couponSvc *services.CouponService,
	emailService *services.EmailService,
	userService *services.UserService,
	settingSvc *services.SettingService,
) *CartStorageHandler {
	return &CartStorageHandler{
		svc:          svc,
		couponSvc:    couponSvc,
		emailService: emailService,
		userService:  userService,
		settingSvc:   settingSvc,
	}
}

func cartUserIDFromCtx(c *gin.Context) (primitive.ObjectID, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(raw.(string))
	if err != nil {
		return primitive.NilObjectID, false
	}
	return id, true
}

// PUT /cart — guarda el carrito del usuario logueado.
func (h *CartStorageHandler) SaveMyCart(c *gin.Context) {
	userID, ok := cartUserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	var req struct {
		Items []domain.CartItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Save(c.Request.Context(), userID, req.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "carrito guardado"})
}

// DELETE /cart — vacía el carrito del usuario logueado.
func (h *CartStorageHandler) ClearMyCart(c *gin.Context) {
	userID, ok := cartUserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
		return
	}
	if err := h.svc.Clear(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "carrito vaciado"})
}

// GET /admin/carts — lista carritos no vacíos enriquecidos con email del usuario.
func (h *CartStorageHandler) ListCarts(c *gin.Context) {
	carts, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type cartRow struct {
		UserID             string            `json:"user_id"`
		Email              string            `json:"email"`
		Items              []domain.CartItem `json:"items"`
		ItemCount          int               `json:"item_count"`
		Total              float64           `json:"total"`
		UpdatedAt          interface{}       `json:"updated_at"`
		LastReminderSentAt interface{}       `json:"last_reminder_sent_at"`
	}

	rows := make([]cartRow, 0, len(carts))
	for _, cart := range carts {
		var total float64
		var count int
		for _, it := range cart.Items {
			total += it.UnitPrice * float64(it.Quantity)
			count += it.Quantity
		}
		email := ""
		if u, err := h.userService.GetUser(c.Request.Context(), cart.UserID); err == nil && u != nil {
			email = u.Email
		}
		var lastReminder interface{}
		if cart.LastReminderSentAt != nil {
			lastReminder = cart.LastReminderSentAt
		}
		rows = append(rows, cartRow{
			UserID:             cart.UserID.Hex(),
			Email:              email,
			Items:              cart.Items,
			ItemCount:          count,
			Total:              total,
			UpdatedAt:          cart.UpdatedAt,
			LastReminderSentAt: lastReminder,
		})
	}
	c.JSON(http.StatusOK, gin.H{"carts": rows})
}

// POST /admin/carts/:userID/send-coupon — envía el email de recuperación.
func (h *CartStorageHandler) SendCoupon(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var req struct {
		CouponID string `json:"coupon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "el carrito está vacío o no existe"})
		return
	}
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	couponID, err := primitive.ObjectIDFromHex(req.CouponID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cupón inválido"})
		return
	}
	coupon, err := h.couponSvc.GetByID(c.Request.Context(), couponID)
	if err != nil || coupon == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cupón no encontrado"})
		return
	}

	body, err := h.settingSvc.GetAbandonedCartBody(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var total float64
	for _, it := range cart.Items {
		total += it.UnitPrice * float64(it.Quantity)
	}

	if err := h.emailService.SendAbandonedCartEmail(user.Email, body, coupon, cart.Items, total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo enviar el email: " + err.Error()})
		return
	}
	_ = h.svc.MarkReminderSent(c.Request.Context(), userID) // no crítico
	c.JSON(http.StatusOK, gin.H{"message": "email enviado a " + user.Email})
}
```

- [ ] **Step 3: Wiring completo en server.go**

Repos:
```go
	cartRepo       := repositories.NewCartRepositoryMongo(db.Collection("carts"))
```
Servicios:
```go
	cartStorageService := services.NewCartStorageService(cartRepo)
```
Handlers:
```go
	cartStorageHandler := handlers.NewCartStorageHandler(cartStorageService, couponService, emailService, userService, settingService)
```
Rutas de cliente (grupo `protected`):
```go
		protected.PUT("/cart", cartStorageHandler.SaveMyCart)
		protected.DELETE("/cart", cartStorageHandler.ClearMyCart)
```
Rutas admin (grupo `admin`):
```go
		admin.GET("/carts", cartStorageHandler.ListCarts)
		admin.POST("/carts/:userID/send-coupon", cartStorageHandler.SendCoupon)
```

- [ ] **Step 4: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 5: Commit**

```bash
git add internal/core/services/cart_storage_service.go internal/adapters/handlers/cart_storage_handler.go internal/app/server.go
git commit -m "feat: servicio, handler y rutas de carritos (cliente y admin)"
```

---

## Task 5: Vaciar carrito server-side en checkout exitoso

**Files:**
- Modify: `internal/adapters/handlers/order_handler.go`

**Interfaces:**
- Consumes: `CartStorageService.Clear` (Task 4). El `OrderHandler` recibe `*services.CartStorageService`.

- [ ] **Step 1: Inyectar CartStorageService en OrderHandler**

En `internal/adapters/handlers/order_handler.go`, agregar el campo al struct y al constructor:
```go
type OrderHandler struct {
	orderService       *services.OrderService
	cartService        *services.CartService
	paymentService     *services.PaymentService
	emailService       *services.EmailService
	couponService      *services.CouponService
	cartStorageService *services.CartStorageService
}

func NewOrderHandler(
	orderService *services.OrderService,
	cartService *services.CartService,
	paymentService *services.PaymentService,
	emailService *services.EmailService,
	couponService *services.CouponService,
	cartStorageService *services.CartStorageService,
) *OrderHandler {
	return &OrderHandler{
		orderService:       orderService,
		cartService:        cartService,
		paymentService:     paymentService,
		emailService:       emailService,
		couponService:      couponService,
		cartStorageService: cartStorageService,
	}
}
```

- [ ] **Step 2: Vaciar el carrito tras crear la orden**

En `Checkout`, después de `go h.emailService.SendOwnerNewOrder(order)` (antes del `c.JSON(http.StatusCreated, ...)`):
```go
	// Vaciar el carrito persistido del usuario: ya concretó la compra.
	if err := h.cartStorageService.Clear(c.Request.Context(), userID); err != nil {
		log.Printf("checkout: no se pudo vaciar el carrito de %s: %v", userID.Hex(), err)
	}
```

- [ ] **Step 3: Actualizar el wiring en server.go**

```go
	orderHandler := handlers.NewOrderHandler(orderService, cartService, paymentService, emailService, couponService, cartStorageService)
```

- [ ] **Step 4: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/handlers/order_handler.go internal/app/server.go
git commit -m "feat: vaciar carrito persistido al concretar checkout"
```

- [ ] **Step 6: Deploy backend** (todas las tareas de backend completas)

```bash
git push coolify v2
git push origin v2
```

---

## Task 6: Sincronización del carrito en el frontend

**Files:**
- Create: `src/lib/cartSync.ts`
- Modify: `src/store/cart.ts`
- Modify: `src/context/AuthContext.tsx`

**Interfaces:**
- Consumes: endpoints `PUT /cart`, `DELETE /cart` (Task 5).
- Produces: `syncCartToServer(items)`, `clearServerCart()` en `cartSync.ts`.

- [ ] **Step 1: Helper de sincronización**

`src/lib/cartSync.ts`:
```ts
import { apiPut, apiDelete } from '@/lib/api';
import type { CartItem } from '@/types/shop';

// Convierte los items del store al formato del backend (CartItem de Go).
function toServerItems(items: CartItem[]) {
  return items.map((i) => ({
    product_id: i.id.split('|')[0],
    variant_sku: i.variant_sku ?? '',
    name: i.name,
    quantity: i.quantity,
    unit_price: i.price,
    item_type: i.type,
  }));
}

function isLoggedIn(): boolean {
  return !!localStorage.getItem('auth_token');
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null;

// Sube el carrito al backend con debounce. Solo si hay sesión.
export function syncCartToServer(items: CartItem[]) {
  if (!isLoggedIn()) return;
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    apiPut('/cart', { items: toServerItems(items) }).catch(() => {
      // silencioso: la sync no debe romper la UX
    });
  }, 2000);
}

// Sube el carrito inmediatamente (ej: al iniciar sesión).
export function pushCartNow(items: CartItem[]) {
  if (!isLoggedIn()) return;
  apiPut('/cart', { items: toServerItems(items) }).catch(() => {});
}

export function clearServerCart() {
  if (!isLoggedIn()) return;
  apiDelete('/cart').catch(() => {});
}
```

> Nota: confirmá los nombres de campos de `CartItem` del store (`id`, `name`, `price`, `quantity`, `variant_sku`, `type`) en `src/types/shop.ts`. Ajustar `toServerItems` si algún nombre difiere.

- [ ] **Step 2: Disparar sync en cambios del store**

En `src/store/cart.ts`, importar el helper arriba:
```ts
import { syncCartToServer } from '@/lib/cartSync';
```
Crear un helper interno que se llame tras cada mutación de `items`. La forma más simple: suscribirse a los cambios al final del archivo, después de crear el store:
```ts
// Sincroniza el carrito con el backend cuando cambian los items (usuarios logueados).
useCart.subscribe((state, prev) => {
  if (state.items !== prev.items) {
    syncCartToServer(state.items);
  }
});
```

- [ ] **Step 3: Subir el carrito al iniciar sesión**

En `src/context/AuthContext.tsx`, tras un login exitoso (donde se setea el token/usuario), importar y llamar:
```ts
import { pushCartNow } from '@/lib/cartSync';
import { useCart } from '@/store/cart';
```
Después de guardar el token en el login (en la función que maneja el login exitoso):
```ts
    // Subir el carrito local al backend al iniciar sesión.
    pushCartNow(useCart.getState().items);
```

> Nota: `useCart.getState()` accede al store fuera de React, es válido con Zustand. Si el token se setea de forma asíncrona, llamar `pushCartNow` después de que `localStorage` ya tenga `auth_token`.

- [ ] **Step 4: Vaciar carrito server-side en checkout**

En `src/pages/CheckoutPage.tsx`, donde ya se llama `clear()` tras el checkout exitoso (línea ~247), agregar:
```ts
import { clearServerCart } from '@/lib/cartSync';
```
Y junto a `clear();`:
```ts
      clear();
      clearServerCart();
```

- [ ] **Step 5: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores.

- [ ] **Step 6: Commit**

```bash
git add src/lib/cartSync.ts src/store/cart.ts src/context/AuthContext.tsx src/pages/CheckoutPage.tsx
git commit -m "feat: sincronizar carrito de usuario logueado con el backend"
```

---

## Task 7: Sección Carritos en el panel (lista + ver detalle)

**Files:**
- Create: `src/pages/admin/CartsPage.tsx`
- Create: `src/components/admin/CartDetailDialog.tsx`
- Modify: `src/App.tsx`
- Modify: `src/pages/admin/AdminLayout.tsx`

**Interfaces:**
- Consumes: `GET /admin/carts` (Task 5).
- Produces: ruta `/admin/carts`, componente `CartDetailDialog`.

- [ ] **Step 1: Dialog de detalle del carrito**

`src/components/admin/CartDetailDialog.tsx`:
```tsx
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { formatPrice } from '@/types/shop';

export interface AdminCartItem {
  product_id: string;
  variant_sku: string;
  name?: string;
  quantity: number;
  unit_price?: number;
  item_type?: string;
}

export interface AdminCart {
  user_id: string;
  email: string;
  items: AdminCartItem[];
  item_count: number;
  total: number;
  updated_at: string;
  last_reminder_sent_at: string | null;
}

export function CartDetailDialog({ cart, onClose }: { cart: AdminCart | null; onClose: () => void }) {
  if (!cart) return null;
  return (
    <Dialog open={!!cart} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Carrito de {cart.email}</DialogTitle>
        </DialogHeader>
        <ul className="divide-y divide-border rounded-lg border border-border overflow-hidden">
          {cart.items.map((item, i) => (
            <li key={i} className="flex items-center justify-between gap-3 px-4 py-3">
              <div className="flex items-center gap-3 min-w-0">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-semibold">
                  {item.quantity}
                </span>
                <span className="truncate text-sm font-medium">
                  {item.name || item.variant_sku || item.product_id.slice(-8)}
                </span>
              </div>
              <span className="text-sm font-semibold shrink-0">
                {formatPrice((item.unit_price ?? 0) * item.quantity)}
              </span>
            </li>
          ))}
        </ul>
        <div className="flex items-center justify-between pt-2">
          <span className="text-sm text-muted-foreground">Total</span>
          <span className="text-lg font-bold">{formatPrice(cart.total)}</span>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Página Carritos**

`src/pages/admin/CartsPage.tsx`:
```tsx
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ShoppingCart, Eye, Mail, FileText } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { apiGet } from '@/lib/api';
import { formatPrice } from '@/types/shop';
import { CartDetailDialog, type AdminCart } from '@/components/admin/CartDetailDialog';

export default function CartsPage() {
  const [viewing, setViewing] = useState<AdminCart | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'carts'],
    queryFn: () => apiGet<{ carts: AdminCart[] }>('/admin/carts').then((r) => r.carts ?? []),
  });

  const carts = data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-primary">Carritos</h1>
          <p className="text-sm text-muted-foreground mt-1">Carritos de clientes que todavía no compraron.</p>
        </div>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Cargando...</p>
      ) : carts.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <ShoppingCart className="h-10 w-10 mx-auto mb-3 opacity-30" />
          <p className="text-sm">No hay carritos activos.</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/40 border-b border-border">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Cliente</th>
                <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Ítems</th>
                <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Total</th>
                <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Actualizado</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-border bg-white">
              {carts.map((cart) => (
                <tr key={cart.user_id} className="hover:bg-muted/20 transition-colors">
                  <td className="px-4 py-3 font-medium">{cart.email || cart.user_id.slice(-8)}</td>
                  <td className="px-4 py-3">{cart.item_count}</td>
                  <td className="px-4 py-3 font-semibold">{formatPrice(cart.total)}</td>
                  <td className="px-4 py-3 text-muted-foreground text-xs">
                    {cart.updated_at ? new Date(cart.updated_at).toLocaleString('es-AR') : '—'}
                    {cart.last_reminder_sent_at && (
                      <div className="text-emerald-600">Recordatorio enviado</div>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <Button variant="outline" size="sm" onClick={() => setViewing(cart)}>
                        <Eye className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <CartDetailDialog cart={viewing} onClose={() => setViewing(null)} />
    </div>
  );
}
```

> Los imports `Mail`, `FileText` y `Button` extra se usan en la Task 8. Si `tsc` marca imports sin usar, agregarlos recién en Task 8.

- [ ] **Step 3: Registrar ruta y link en sidebar**

En `src/App.tsx`:
```tsx
import AdminCartsPage from './pages/admin/CartsPage.tsx';
```
Dentro del grupo de rutas admin:
```tsx
                <Route path="carts" element={<AdminCartsPage />} />
```
En `src/pages/admin/AdminLayout.tsx`, agregar al import de iconos `ShoppingCart` y a `TABS`:
```tsx
  { to: '/admin/carts',    label: 'Carritos',  icon: ShoppingCart },
```

- [ ] **Step 4: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores (quitar imports no usados si los marca).

- [ ] **Step 5: Commit**

```bash
git add src/pages/admin/CartsPage.tsx src/components/admin/CartDetailDialog.tsx src/App.tsx src/pages/admin/AdminLayout.tsx
git commit -m "feat: seccion Carritos en el panel admin"
```

---

## Task 8: Enviar cupón desde el carrito (dropdown + preview + envío)

**Files:**
- Create: `src/components/admin/SendCouponDialog.tsx`
- Modify: `src/pages/admin/CartsPage.tsx`

**Interfaces:**
- Consumes: `GET /admin/coupons`, `GET /admin/settings/abandoned-cart-email`, `POST /admin/carts/:userID/send-coupon`.

- [ ] **Step 1: Dialog de envío de cupón con preview**

`src/components/admin/SendCouponDialog.tsx`:
```tsx
import { useMemo, useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { apiGet, apiPost } from '@/lib/api';
import { formatPrice } from '@/types/shop';
import { toast } from 'sonner';
import type { AdminCart } from './CartDetailDialog';

interface Coupon {
  id: string;
  code: string;
  discount_percent: number;
  active: boolean;
  expires_at?: string;
}

export function SendCouponDialog({ cart, onClose }: { cart: AdminCart | null; onClose: () => void }) {
  const [couponId, setCouponId] = useState('');

  const { data: coupons } = useQuery({
    queryKey: ['admin', 'coupons'],
    queryFn: () => apiGet<{ coupons: Coupon[] }>('/admin/coupons').then((r) => r.coupons ?? []),
  });
  const { data: template } = useQuery({
    queryKey: ['admin', 'abandoned-template'],
    queryFn: () => apiGet<{ body: string }>('/admin/settings/abandoned-cart-email').then((r) => r.body),
  });

  const selected = (coupons ?? []).find((c) => c.id === couponId);

  const preview = useMemo(() => {
    if (!template || !cart) return '';
    const productos = cart.items
      .map((i) => `- ${i.name || i.variant_sku} x${i.quantity}`)
      .join('\n');
    const vencimiento = selected?.expires_at
      ? new Date(selected.expires_at).toLocaleDateString('es-AR')
      : 'sin vencimiento';
    return template
      .replaceAll('{{codigo}}', selected?.code ?? '{{codigo}}')
      .replaceAll('{{descuento}}', selected ? `${selected.discount_percent}%` : '{{descuento}}')
      .replaceAll('{{vencimiento}}', vencimiento)
      .replaceAll('{{productos}}', productos)
      .replaceAll('{{total}}', formatPrice(cart.total));
  }, [template, cart, selected]);

  const send = useMutation({
    mutationFn: () => apiPost(`/admin/carts/${cart!.user_id}/send-coupon`, { coupon_id: couponId }),
    onSuccess: () => { toast.success('Email enviado'); onClose(); },
    onError: (e: Error) => toast.error(e.message),
  });

  const activeCoupons = (coupons ?? []).filter((c) => c.active);

  return (
    <Dialog open={!!cart} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Enviar cupón a {cart?.email}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label className="text-xs">Cupón</Label>
            <Select value={couponId} onValueChange={setCouponId}>
              <SelectTrigger><SelectValue placeholder="Elegí un cupón" /></SelectTrigger>
              <SelectContent>
                {activeCoupons.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.code} — {c.discount_percent}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs">Vista previa</Label>
            <pre className="whitespace-pre-wrap rounded-lg border border-border bg-muted/30 p-3 text-xs text-foreground max-h-64 overflow-y-auto">
              {preview || 'Elegí un cupón para ver el mensaje.'}
            </pre>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Cancelar</Button>
          <Button onClick={() => send.mutate()} disabled={!couponId || send.isPending}>
            {send.isPending ? 'Enviando...' : 'Enviar email'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Botón "Enviar cupón" en la tabla de Carritos**

En `src/pages/admin/CartsPage.tsx`, importar y agregar estado:
```tsx
import { SendCouponDialog } from '@/components/admin/SendCouponDialog';
```
```tsx
  const [sending, setSending] = useState<AdminCart | null>(null);
```
En la celda de acciones, junto al botón de ver:
```tsx
                      <Button variant="outline" size="sm" onClick={() => setSending(cart)}>
                        <Mail className="h-3.5 w-3.5" />
                      </Button>
```
Antes del cierre del componente, junto a `CartDetailDialog`:
```tsx
      <SendCouponDialog cart={sending} onClose={() => setSending(null)} />
```

- [ ] **Step 3: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores.

- [ ] **Step 4: Commit**

```bash
git add src/components/admin/SendCouponDialog.tsx src/pages/admin/CartsPage.tsx
git commit -m "feat: enviar cupon de recuperacion desde el carrito con preview"
```

---

## Task 9: Editor del template del email

**Files:**
- Create: `src/components/admin/AbandonedEmailTemplateDialog.tsx`
- Modify: `src/pages/admin/CartsPage.tsx`

**Interfaces:**
- Consumes: `GET/PUT /admin/settings/abandoned-cart-email`.

- [ ] **Step 1: Dialog editor del template**

`src/components/admin/AbandonedEmailTemplateDialog.tsx`:
```tsx
import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { apiGet, apiPut } from '@/lib/api';
import { toast } from 'sonner';

const VARIABLES = ['{{codigo}}', '{{descuento}}', '{{vencimiento}}', '{{productos}}', '{{total}}'];

export function AbandonedEmailTemplateDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const qc = useQueryClient();
  const [body, setBody] = useState('');
  const [defaultBody, setDefaultBody] = useState('');

  const { data } = useQuery({
    queryKey: ['admin', 'abandoned-template-full'],
    queryFn: () => apiGet<{ body: string; default_body: string }>('/admin/settings/abandoned-cart-email'),
    enabled: open,
  });

  useEffect(() => {
    if (data) { setBody(data.body); setDefaultBody(data.default_body); }
  }, [data]);

  const save = useMutation({
    mutationFn: () => apiPut('/admin/settings/abandoned-cart-email', { body }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'abandoned-template'] });
      qc.invalidateQueries({ queryKey: ['admin', 'abandoned-template-full'] });
      toast.success('Plantilla guardada');
      onClose();
    },
    onError: (e: Error) => toast.error(e.message),
  });

  const insertVar = (v: string) => setBody((b) => b + ' ' + v);

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Plantilla del email de recuperación</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div className="flex flex-wrap gap-1.5">
            {VARIABLES.map((v) => (
              <button
                key={v}
                type="button"
                onClick={() => insertVar(v)}
                className="rounded-md border border-border bg-muted/40 px-2 py-1 font-mono text-xs hover:bg-muted"
              >
                {v}
              </button>
            ))}
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs">Mensaje</Label>
            <Textarea rows={10} value={body} onChange={(e) => setBody(e.target.value)} />
          </div>
          <button
            type="button"
            onClick={() => setBody(defaultBody)}
            className="text-xs text-muted-foreground underline"
          >
            Restaurar mensaje por defecto
          </button>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Cancelar</Button>
          <Button onClick={() => save.mutate()} disabled={save.isPending}>
            {save.isPending ? 'Guardando...' : 'Guardar'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

> Verificá que exista `src/components/ui/textarea.tsx`. Si no, crear el componente shadcn Textarea estándar o reemplazar por un `<textarea>` con clases de `Input`.

- [ ] **Step 2: Botón "Editar plantilla" en la página Carritos**

En `src/pages/admin/CartsPage.tsx`, importar y agregar estado + botón en el header:
```tsx
import { AbandonedEmailTemplateDialog } from '@/components/admin/AbandonedEmailTemplateDialog';
```
```tsx
  const [editTemplate, setEditTemplate] = useState(false);
```
En el header, junto al título (en el `flex items-center justify-between`):
```tsx
        <Button variant="outline" onClick={() => setEditTemplate(true)}>
          <FileText className="h-4 w-4 mr-2" /> Editar plantilla del mail
        </Button>
```
Junto a los otros diálogos:
```tsx
      <AbandonedEmailTemplateDialog open={editTemplate} onClose={() => setEditTemplate(false)} />
```

- [ ] **Step 3: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores.

- [ ] **Step 4: Commit y deploy frontend**

```bash
git add src/components/admin/AbandonedEmailTemplateDialog.tsx src/pages/admin/CartsPage.tsx
git commit -m "feat: editor de plantilla del email de recuperacion"
git push origin v2
```

---

## Task 10: Prueba end-to-end manual

**Files:** ninguno (verificación).

- [ ] **Step 1: Reiniciar el backend local** con el código nuevo (`go run ./cmd/api` o docker compose rebuild).

- [ ] **Step 2: Sync del carrito**
  1. Logueate en la tienda como cliente.
  2. Agregá productos al carrito.
  3. En el panel admin → Carritos, tiene que aparecer tu carrito con ítems, total y fecha.

- [ ] **Step 3: Carrito se vacía al comprar**
  1. Con ese carrito, hacé un checkout (transferencia).
  2. En Carritos, ese carrito debe desaparecer (se vació server-side).

- [ ] **Step 4: Enviar cupón**
  1. Armá otro carrito logueado.
  2. En Carritos → "Enviar cupón" → elegí un cupón activo → verificá el preview con las variables rellenadas → Enviar.
  3. Verificá que llegue el email (Resend) y que en la fila aparezca "Recordatorio enviado".

- [ ] **Step 5: Editar plantilla**
  1. Carritos → "Editar plantilla del mail".
  2. Cambiá el texto usando las variables, guardá.
  3. Reabrí "Enviar cupón" y confirmá que el preview usa el texto nuevo.
  4. "Restaurar mensaje por defecto" y guardar vuelve al original.

---

## Self-Review Notes

- **Spec coverage:** persistencia carrito (Task 1, 4), settings/template (Task 2), email de recuperación (Task 3), endpoints admin de carritos (Task 4), vaciar en checkout (Task 5), sync frontend (Task 6), vista panel (Task 7), envío con preview (Task 8), editor template (Task 9), prueba manual (Task 10). ✅
- **Orden de dependencias backend:** cada task de backend (1→5) deja el `server.go` compilando por sí sola. El orden importa: Task 2 (settings) y Task 3 (email + GetByID) van antes de Task 4 (handler de carrito) porque el handler los consume. Ejecutar 1→5 en orden.
- **`GetByID` de cupón:** verificar si ya existe en `coupon_repository_mongo.go` antes de agregarlo (Task 3, Step 2) para no duplicar.
- **Nombres de campos del store (`CartItem` frontend):** confirmar en `src/types/shop.ts` (`id`, `name`, `price`, `quantity`, `variant_sku`, `type`) antes de Task 6.
- **Componente `Textarea`:** verificar que exista `src/components/ui/textarea.tsx` antes de Task 9; si no, crearlo o usar `<textarea>` con clases de `Input`.
