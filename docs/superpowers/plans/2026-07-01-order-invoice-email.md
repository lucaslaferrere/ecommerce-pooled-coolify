# Order Invoice Email Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Que el admin suba un PDF de factura a un pedido y se lo envíe al cliente por email (adjunto), quedando la factura guardada en la base para reenviarla.

**Architecture:** El PDF se guarda como base64 en una colección `invoices` (una por pedido), no en el filesystem (para sobrevivir redeploys de Coolify). El pedido guarda solo una marca liviana `invoice_sent_at`. El email se envía vía Resend con el PDF adjunto, solo al cliente. Endpoints admin para subir+enviar y reenviar.

**Tech Stack:** Go + Gin + MongoDB (backend `ecommerce-pooled`), React + TypeScript + Vite + React Query (frontend `poolside-commerce-hub`), Resend para email.

## Global Constraints

- Backend production deploy: `git push coolify v2` (repo `ecommerce-pooled-coolify`) + `git push origin v2`. Frontend: `git push origin v2`.
- Nunca pushear `.env`.
- Verificación por tarea: backend `go build ./...`; frontend `npx tsc --noEmit`; + prueba manual. No hay tests unitarios en el proyecto.
- Commits: conventional commits, sin atribución AI / Co-Authored-By.
- El PDF se guarda en MongoDB como base64, NO en filesystem.
- El email de factura se envía SOLO a `order.CustomerEmail`, NUNCA a `OwnerEmails`.
- Solo `.pdf`, máximo 8 MB.
- UI en español; identificadores en inglés.

---

## File Structure

**Backend (`ecommerce-pooled`):**
- Create: `internal/core/domain/invoice.go` — entidad `Invoice`.
- Create: `internal/adapters/repositories/invoice_repository_mongo.go` — upsert/get por order_id.
- Modify: `internal/core/domain/order.go` — campo `InvoiceSentAt *time.Time`.
- Modify: `internal/core/services/order_service.go` — `SetInvoiceSentAt`.
- Modify: `internal/core/services/email_service.go` — `SendInvoice` + helper `postResend`.
- Create: `internal/core/services/invoice_service.go` — orquesta guardar + enviar + reenviar.
- Create: `internal/adapters/handlers/invoice_handler.go` — endpoints admin (multipart).
- Modify: `internal/app/server.go` — wiring + rutas.

**Frontend (`poolside-commerce-hub`):**
- Modify: `src/types/shop.ts` — `invoice_sent_at?` en `Order`.
- Modify: `src/components/admin/OrderDetailDialog.tsx` — sección "Factura" (subir/enviar/estado/reenviar).
- Modify: `src/components/admin/OrdersTable.tsx` — indicador "Factura enviada".

---

## Task 1: Colección invoices + campo en Order + marcado

**Files:**
- Create: `internal/core/domain/invoice.go`
- Create: `internal/adapters/repositories/invoice_repository_mongo.go`
- Modify: `internal/core/domain/order.go`
- Modify: `internal/core/services/order_service.go`

**Interfaces:**
- Produces: `domain.Invoice{ OrderID, Filename, Content(base64), SentAt *time.Time, CreatedAt, UpdatedAt }`
- Produces: `InvoiceRepositoryMongo` con `Upsert(ctx, inv) error`, `GetByOrderID(ctx, orderID) (*Invoice, error)`, `SetSentAt(ctx, orderID, t) error`.
- Produces: `OrderService.SetInvoiceSentAt(ctx, id, t) error`.
- Produces: `Order.InvoiceSentAt *time.Time`.

- [ ] **Step 1: Entidad Invoice**

`internal/core/domain/invoice.go`:
```go
package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invoice es la factura (PDF) asociada a un pedido. El PDF se guarda como base64
// en Content para no depender del filesystem (que se pierde en redeploys).
type Invoice struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id,omitempty"`
	OrderID   primitive.ObjectID `bson:"order_id"        json:"order_id"`
	Filename  string             `bson:"filename"        json:"filename"`
	Content   string             `bson:"content"         json:"-"` // base64; nunca se expone por JSON
	SentAt    *time.Time         `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"      json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"      json:"updated_at"`
}
```

- [ ] **Step 2: Repositorio Mongo**

`internal/adapters/repositories/invoice_repository_mongo.go`:
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

type InvoiceRepositoryMongo struct {
	collection *mongo.Collection
}

func NewInvoiceRepositoryMongo(col *mongo.Collection) *InvoiceRepositoryMongo {
	return &InvoiceRepositoryMongo{collection: col}
}

// Upsert reemplaza la factura del pedido (una por order_id).
func (r *InvoiceRepositoryMongo) Upsert(ctx context.Context, inv *domain.Invoice) error {
	now := time.Now()
	inv.UpdatedAt = now
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"order_id": inv.OrderID},
		bson.M{
			"$set": bson.M{
				"filename":   inv.Filename,
				"content":    inv.Content,
				"sent_at":    inv.SentAt,
				"updated_at": now,
			},
			"$setOnInsert": bson.M{"order_id": inv.OrderID, "created_at": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *InvoiceRepositoryMongo) GetByOrderID(ctx context.Context, orderID primitive.ObjectID) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := r.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&inv)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &inv, err
}

// SetSentAt marca la fecha de envío de la factura del pedido.
func (r *InvoiceRepositoryMongo) SetSentAt(ctx context.Context, orderID primitive.ObjectID, t time.Time) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"order_id": orderID},
		bson.M{"$set": bson.M{"sent_at": t, "updated_at": time.Now()}},
	)
	return err
}
```

- [ ] **Step 3: Campo InvoiceSentAt en Order**

En `internal/core/domain/order.go`, agregar el campo junto a los otros opcionales (por ejemplo después de `TrackingNumber`):
```go
	InvoiceSentAt   *time.Time         `bson:"invoice_sent_at,omitempty" json:"invoice_sent_at,omitempty"`
```

- [ ] **Step 4: OrderService.SetInvoiceSentAt**

En `internal/core/services/order_service.go`, agregar (mismo patrón que `SetTracking`):
```go
// SetInvoiceSentAt marca en el pedido la fecha de envío de la factura.
func (s *OrderService) SetInvoiceSentAt(ctx context.Context, id primitive.ObjectID, t time.Time) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	order.InvoiceSentAt = &t
	order.UpdatedAt = time.Now()
	return s.orderRepository.Update(ctx, order)
}
```

- [ ] **Step 5: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila, sin salida.

- [ ] **Step 6: Commit**

```bash
git add internal/core/domain/invoice.go internal/adapters/repositories/invoice_repository_mongo.go internal/core/domain/order.go internal/core/services/order_service.go
git commit -m "feat: entidad Invoice, repositorio y marca invoice_sent_at en pedido"
```

---

## Task 2: EmailService.SendInvoice con adjunto

**Files:**
- Modify: `internal/core/services/email_service.go`

**Interfaces:**
- Produces: `EmailService.SendInvoice(order *domain.Order, pdfBase64, filename string) error`.

Método exportado nuevo: compila aunque todavía no se use (lo consume Task 3).

- [ ] **Step 1: Factorizar el POST a Resend en un helper y agregar SendInvoice**

En `internal/core/services/email_service.go`. El método `Send` actual hace marshal del payload y POST a Resend. Extraer el POST a un helper `postResend(payload interface{}) error` y reusarlo desde `Send` y `SendInvoice`. Concretamente:

Reemplazar el cuerpo de `Send` (desde el armado del payload hasta el return) para que delegue en el helper, y agregar el helper + `SendInvoice`. El `Send` mantiene su firma pública `Send(to, subject, html string) error`.

```go
// postResend serializa el payload y lo envía a la API de Resend.
func (s *EmailService) postResend(payload interface{}) error {
	if s.apiKey == "" {
		return fmt.Errorf("servicio de email no configurado")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend error: status %d", resp.StatusCode)
	}
	return nil
}
```

Y hacer que `Send` use el helper:
```go
// Send envía un email HTML al destinatario indicado.
func (s *EmailService) Send(to, subject, html string) error {
	payload := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		HTML    string   `json:"html"`
	}{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
	return s.postResend(payload)
}
```

- [ ] **Step 2: Agregar SendInvoice**

En el mismo archivo:
```go
// SendInvoice envía la factura (PDF adjunto) SOLO al cliente del pedido.
// pdfBase64 es el contenido del PDF en base64; filename el nombre del adjunto.
func (s *EmailService) SendInvoice(order *domain.Order, pdfBase64, filename string) error {
	if !s.Enabled() {
		return nil
	}
	ref := strings.ToUpper(order.ID.Hex()[len(order.ID.Hex())-8:])
	subject := fmt.Sprintf("Factura de tu pedido #%s — Pooled", ref)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="es"><head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:-apple-system,Segoe UI,Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 0;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">
        <tr><td style="background:#0B1F3A;padding:24px 32px;">
          <h1 style="margin:0;color:#fff;font-size:18px;">Tu factura</h1>
        </td></tr>
        <tr><td style="padding:28px 32px;color:#374151;font-size:14px;line-height:1.6;">
          Adjuntamos la factura de tu pedido <strong>#%s</strong>.<br><br>
          ¡Gracias por tu compra!
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`, ref)

	payload := struct {
		From        string              `json:"from"`
		To          []string            `json:"to"`
		Subject     string              `json:"subject"`
		HTML        string              `json:"html"`
		Attachments []map[string]string `json:"attachments"`
	}{
		From:    s.from,
		To:      []string{order.CustomerEmail},
		Subject: subject,
		HTML:    html,
		Attachments: []map[string]string{
			{"filename": filename, "content": pdfBase64},
		},
	}
	return s.postResend(payload)
}
```

- [ ] **Step 3: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 4: Commit**

```bash
git add internal/core/services/email_service.go
git commit -m "feat: SendInvoice con PDF adjunto via Resend"
```

---

## Task 3: InvoiceService + handler + rutas

**Files:**
- Create: `internal/core/services/invoice_service.go`
- Create: `internal/adapters/handlers/invoice_handler.go`
- Modify: `internal/app/server.go`

**Interfaces:**
- Consumes: `InvoiceRepositoryMongo` (Task 1), `OrderService.GetOrder`/`SetInvoiceSentAt` (existente/Task 1), `EmailService.SendInvoice` (Task 2).
- Produces: `InvoiceService.SaveAndSend(ctx, orderID, filename, pdfBase64) (time.Time, error)`, `.Resend(ctx, orderID) (time.Time, error)`.
- Produces: rutas `POST /admin/orders/:id/invoice`, `POST /admin/orders/:id/invoice/resend`.

- [ ] **Step 1: InvoiceService**

`internal/core/services/invoice_service.go`:
```go
package services

import (
	"context"
	"errors"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvoiceNotFound = errors.New("factura no encontrada")

type InvoiceService struct {
	repo         *repositories.InvoiceRepositoryMongo
	orderService *OrderService
	emailService *EmailService
}

func NewInvoiceService(repo *repositories.InvoiceRepositoryMongo, orderService *OrderService, emailService *EmailService) *InvoiceService {
	return &InvoiceService{repo: repo, orderService: orderService, emailService: emailService}
}

// SaveAndSend guarda (o reemplaza) el PDF de la factura y lo envía al cliente.
// Primero persiste el PDF; si el envío falla, la factura queda guardada para
// reintentar con Resend. Solo marca sent_at cuando el envío fue exitoso.
func (s *InvoiceService) SaveAndSend(ctx context.Context, orderID primitive.ObjectID, filename, pdfBase64 string) (time.Time, error) {
	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if order == nil {
		return time.Time{}, ErrOrderNotFound
	}

	// Guardar/reemplazar el PDF (sin sent_at todavía).
	inv := &domain.Invoice{OrderID: orderID, Filename: filename, Content: pdfBase64, SentAt: nil}
	if err := s.repo.Upsert(ctx, inv); err != nil {
		return time.Time{}, err
	}

	// Enviar. Si falla, la factura queda guardada.
	if err := s.emailService.SendInvoice(order, pdfBase64, filename); err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	_ = s.repo.SetSentAt(ctx, orderID, now)
	_ = s.orderService.SetInvoiceSentAt(ctx, orderID, now)
	return now, nil
}

// Resend reenvía la factura ya guardada del pedido.
func (s *InvoiceService) Resend(ctx context.Context, orderID primitive.ObjectID) (time.Time, error) {
	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if order == nil {
		return time.Time{}, ErrOrderNotFound
	}
	inv, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if inv == nil {
		return time.Time{}, ErrInvoiceNotFound
	}
	if err := s.emailService.SendInvoice(order, inv.Content, inv.Filename); err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	_ = s.repo.SetSentAt(ctx, orderID, now)
	_ = s.orderService.SetInvoiceSentAt(ctx, orderID, now)
	return now, nil
}
```

> Nota: verificá que `OrderService.GetOrder(ctx, id) (*domain.Order, error)` exista (se usa en otros handlers). Si el nombre difiere (p. ej. `GetByID`), ajustá la llamada. `ErrOrderNotFound` ya está definido en el paquete services.

- [ ] **Step 2: Handler (multipart, validación PDF + tamaño)**

`internal/adapters/handlers/invoice_handler.go`:
```go
package handlers

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxInvoiceBytes = 8 * 1024 * 1024 // 8 MB

type InvoiceHandler struct {
	svc *services.InvoiceService
}

func NewInvoiceHandler(svc *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// POST /admin/orders/:id/invoice — multipart con campo "file" (PDF).
func (h *InvoiceHandler) Upload(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo"})
		return
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el archivo debe ser un PDF"})
		return
	}
	if fileHeader.Size > maxInvoiceBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el archivo supera los 8 MB"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el archivo"})
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el archivo"})
		return
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(raw)

	sentAt, err := h.svc.SaveAndSend(c.Request.Context(), orderID, fileHeader.Filename, pdfBase64)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pedido no encontrado"})
			return
		}
		// La factura pudo guardarse aunque el envío falle.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "la factura se guardó pero no se pudo enviar: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoice_sent_at": sentAt})
}

// POST /admin/orders/:id/invoice/resend — reenvía la factura guardada.
func (h *InvoiceHandler) Resend(c *gin.Context) {
	orderID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	sentAt, err := h.svc.Resend(c.Request.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvoiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "este pedido no tiene factura cargada"})
		case errors.Is(err, services.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "pedido no encontrado"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo reenviar: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoice_sent_at": sentAt})
}
```

- [ ] **Step 3: Wiring en server.go**

Repos:
```go
	invoiceRepo    := repositories.NewInvoiceRepositoryMongo(db.Collection("invoices"))
```
Servicios:
```go
	invoiceService := services.NewInvoiceService(invoiceRepo, orderService, emailService)
```
Handlers:
```go
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
```
Rutas admin (dentro del grupo `admin`, junto a las otras rutas de orders):
```go
		admin.POST("/orders/:id/invoice", invoiceHandler.Upload)
		admin.POST("/orders/:id/invoice/resend", invoiceHandler.Resend)
```

> Nota: el grupo admin ya tiene `admin.PATCH("/orders/:id/status", ...)` y `admin.DELETE("/orders/:id", ...)`. Gin permite `:id` compartido; agregá estas dos rutas en el mismo bloque.

- [ ] **Step 4: Compilar**

Run: `cd C:/Users/lafer/ecommerce-pooled && go build ./...`
Expected: compila.

- [ ] **Step 5: Commit**

```bash
git add internal/core/services/invoice_service.go internal/adapters/handlers/invoice_handler.go internal/app/server.go
git commit -m "feat: endpoints admin para enviar y reenviar factura"
```

---

## Task 4: Sección Factura en el detalle del pedido (frontend)

**Files:**
- Modify: `src/types/shop.ts`
- Modify: `src/components/admin/OrderDetailDialog.tsx`

**Interfaces:**
- Consumes: `POST /admin/orders/:id/invoice` (multipart `file`), `POST /admin/orders/:id/invoice/resend`.

- [ ] **Step 1: Agregar invoice_sent_at al tipo Order**

En `src/types/shop.ts`, dentro de `interface Order`, junto a `tracking_number`:
```ts
  invoice_sent_at?: string;
```

- [ ] **Step 2: Sección "Factura" en OrderDetailDialog**

En `src/components/admin/OrderDetailDialog.tsx`. Agregar imports arriba:
```tsx
import { useEffect, useRef, useState } from 'react';
import { apiPostForm, apiPost } from '@/lib/api';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { FileText } from 'lucide-react';
```
> Si `FileText` ya está importado de lucide-react en este archivo, no lo dupliques. Si el archivo ya importa algo de `react`, sumá `useEffect, useRef, useState` a esa línea en vez de duplicar.

**IMPORTANTE — ubicación de los hooks:** el componente tiene un early return `if (!order) return null;` cerca del inicio y se monta de forma persistente (con `order` en `null` hasta abrir un pedido). Los hooks DEBEN ir ANTES de ese `if (!order) return null;` (nunca después — romperían las reglas de hooks), y el estado se sincroniza con el pedido actual vía `useEffect` (si no, queda stale al abrir otro pedido). Colocá este bloque como las PRIMERAS líneas del cuerpo del componente, arriba del `if (!order) return null;`:
```tsx
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);
  const [invoiceSentAt, setInvoiceSentAt] = useState<string | undefined>(undefined);
  const [uploading, setUploading] = useState(false);

  // Sincronizar el estado local con el pedido actual (evita estado stale entre pedidos).
  useEffect(() => {
    setInvoiceSentAt(order?.invoice_sent_at);
  }, [order?.id, order?.invoice_sent_at]);

  const uploadInvoice = async (file: File) => {
    if (!order) return;
    if (file.type !== 'application/pdf') { toast.error('El archivo debe ser un PDF'); return; }
    if (file.size > 8 * 1024 * 1024) { toast.error('El archivo supera los 8 MB'); return; }
    setUploading(true);
    try {
      const form = new FormData();
      form.append('file', file);
      const res = await apiPostForm<{ invoice_sent_at: string }>(`/admin/orders/${order.id}/invoice`, form);
      setInvoiceSentAt(res.invoice_sent_at);
      qc.invalidateQueries({ queryKey: ['admin', 'orders'] });
      toast.success('Factura enviada al cliente');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'No se pudo enviar la factura');
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  const resendInvoice = async () => {
    if (!order) return;
    setUploading(true);
    try {
      const res = await apiPost<{ invoice_sent_at: string }>(`/admin/orders/${order.id}/invoice/resend`, {});
      setInvoiceSentAt(res.invoice_sent_at);
      qc.invalidateQueries({ queryKey: ['admin', 'orders'] });
      toast.success('Factura reenviada');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'No se pudo reenviar');
    } finally {
      setUploading(false);
    }
  };
```

> Verificá la key real que usa el listado de pedidos del admin en `useAdminOrders` (probablemente `['admin', 'orders', ...]`). Usá el prefijo correcto en `invalidateQueries` para que la fila refleje el estado. Si la key difiere, ajustá `['admin', 'orders']`.

Agregar la sección visual dentro del cuerpo del diálogo (por ejemplo después de la sección "Factura A" o antes del bloque de Productos), siguiendo el estilo de `SectionTitle` ya usado en el archivo:
```tsx
            {/* Factura (PDF) */}
            <section className="mt-4 flex flex-col gap-2 rounded-xl border border-border bg-muted/30 p-4">
              <SectionTitle icon={FileText}>Factura</SectionTitle>
              {invoiceSentAt ? (
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm text-emerald-600 font-medium">
                    Factura enviada el {new Date(invoiceSentAt).toLocaleString('es-AR')}
                  </p>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={resendInvoice}
                      disabled={uploading}
                      className="text-xs rounded-md border border-border px-3 py-1.5 hover:bg-muted disabled:opacity-50"
                    >
                      Reenviar
                    </button>
                    <button
                      type="button"
                      onClick={() => fileRef.current?.click()}
                      disabled={uploading}
                      className="text-xs rounded-md border border-border px-3 py-1.5 hover:bg-muted disabled:opacity-50"
                    >
                      Subir otra
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm text-muted-foreground">Todavía no se envió factura para este pedido.</p>
                  <button
                    type="button"
                    onClick={() => fileRef.current?.click()}
                    disabled={uploading}
                    className="text-xs rounded-md bg-primary text-white px-3 py-1.5 hover:opacity-90 disabled:opacity-50"
                  >
                    {uploading ? 'Enviando...' : 'Enviar factura al cliente'}
                  </button>
                </div>
              )}
              <input
                ref={fileRef}
                type="file"
                accept="application/pdf"
                className="hidden"
                onChange={(e) => { const f = e.target.files?.[0]; if (f) uploadInvoice(f); }}
              />
            </section>
```

- [ ] **Step 3: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores.

- [ ] **Step 4: Commit**

```bash
git add src/types/shop.ts src/components/admin/OrderDetailDialog.tsx
git commit -m "feat: seccion Factura en el detalle del pedido (subir, enviar, reenviar)"
```

---

## Task 5: Indicador "Factura enviada" en la tabla de pedidos

**Files:**
- Modify: `src/components/admin/OrdersTable.tsx`

**Interfaces:**
- Consumes: `Order.invoice_sent_at` (Task 4).

- [ ] **Step 1: Mostrar el indicador**

En `src/components/admin/OrdersTable.tsx`, leer el archivo primero para ubicar dónde se renderiza cada fila y qué columnas hay. Agregar un badge/indicador cuando `order.invoice_sent_at` esté presente, siguiendo el estilo de badges ya usado en la tabla. Ejemplo de elemento a insertar en la celda apropiada (p. ej. junto al estado o el tracking):
```tsx
{order.invoice_sent_at && (
  <span className="inline-flex items-center gap-1 text-[11px] text-emerald-600">
    <FileText className="h-3 w-3" /> Factura enviada
  </span>
)}
```
Importar `FileText` de `lucide-react` si no está ya importado en el archivo.

> Si `OrdersTable` no recibe el campo por su tipo, confirmá que la fila usa el tipo `Order` (que ya tiene `invoice_sent_at` desde Task 4). Ajustá la ubicación del indicador a la estructura real de la tabla.

- [ ] **Step 2: Typecheck**

Run: `cd C:/Users/lafer/poolside-commerce-hub && npx tsc --noEmit`
Expected: sin errores.

- [ ] **Step 3: Commit y deploy**

```bash
git add src/components/admin/OrdersTable.tsx
git commit -m "feat: indicador de factura enviada en la tabla de pedidos"
```
Deploy (tras aprobar el review final):
- Backend: `git push coolify v2 && git push origin v2`
- Frontend: `git push origin v2`

---

## Task 6: Prueba end-to-end manual

**Files:** ninguno (verificación).

- [ ] **Step 1: Reiniciar el backend local** con el código nuevo.
- [ ] **Step 2: Enviar factura**
  1. Admin → un pedido → detalle → sección Factura → "Enviar factura al cliente" → elegí un PDF.
  2. Verificá que llegue el mail al **cliente** (no a los owners) con el PDF adjunto.
  3. El detalle muestra "Factura enviada el {fecha}"; la fila de la tabla muestra el indicador.
- [ ] **Step 3: Reenviar** desde el detalle → llega de nuevo, no pide resubir.
- [ ] **Step 4: Reemplazar** ("Subir otra") con otro PDF → se manda el nuevo.
- [ ] **Step 5: Validaciones** → subir un archivo que no sea PDF o > 8 MB → error claro, no se envía.

---

## Self-Review Notes

- **Spec coverage:** persistencia en Mongo colección invoices (Task 1); marca liviana invoice_sent_at en Order (Task 1); envío con adjunto solo al cliente (Task 2); guardar-primero-enviar-después + reenvío (Task 3); UI subir/enviar/estado/reenviar (Task 4); indicador en tabla (Task 5); prueba manual (Task 6). ✅
- **Solo cliente, no owners:** `SendInvoice` usa `order.CustomerEmail` únicamente (Task 2). Verificado contra el constraint.
- **Orden de dependencias backend:** Task 1 y 2 son independientes y compilan solas; Task 3 las consume. Ejecutar 1→3 antes de probar.
- **Nombre de método OrderService:** el plan usa `GetOrder`; verificar que exista (se usa en `order_handler.go`); si es `GetByID`, ajustar en `invoice_service.go`.
- **React Query key del listado de pedidos:** confirmar el prefijo real en `useAdminOrders` antes de Task 4/5 (`['admin','orders']` asumido).
