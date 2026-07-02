# Order Invoice Email — Design

**Date:** 2026-07-01
**Status:** Approved (pending final spec review)
**Scope:** Admin sube un PDF de factura a un pedido y se lo envía al cliente por email, adjunto. La factura queda guardada para reenviarla.

## Objetivo

Desde el detalle de un pedido en el admin, el operador sube el PDF de la factura,
lo envía al email del cliente con el PDF adjunto (estilo Mercado Libre), y la
factura queda persistida para poder reenviarla sin volver a subirla.

## Decisiones tomadas

1. **Guardar la factura** asociada al pedido (no solo enviar).
2. **Solo admin** en esta versión. El cliente recibe la factura por mail; NO se
   agrega descarga desde el sitio del cliente (queda para más adelante).
3. **Persistencia en MongoDB**, no en filesystem. El filesystem local del
   contenedor (`./uploads`) se pierde en redeploys de Coolify salvo volumen
   persistente; una factura no puede depender de eso. El PDF se guarda como base64
   en una colección dedicada.
4. **Un PDF por pedido**, reemplazable (subir uno nuevo pisa el anterior).
5. **Solo `.pdf`**, máximo ~8 MB (para entrar holgado en el límite de 16 MB por
   documento de Mongo tras el ~33% de overhead de base64).
6. Texto del mail **fijo** (no editable) en esta versión.
7. Estado **"Factura enviada" + fecha** visible en el detalle y en la fila del pedido.

## Arquitectura

### Persistencia

**Colección `invoices`** (una por pedido):
- `order_id` (ObjectID, único)
- `filename` (string — nombre original del PDF)
- `content` (string — PDF en base64)
- `sent_at` (time)
- `created_at` / `updated_at`

**Campo liviano en `Order`:** `InvoiceSentAt *time.Time` (`bson:"invoice_sent_at,omitempty"`).
Permite mostrar el estado "Factura enviada" en el listado y detalle de pedidos sin
cargar el PDF pesado. La colección `invoices` guarda el binario; el `Order` solo
guarda la marca de tiempo.

Racional de la separación: las consultas de pedidos (listado admin, historial del
cliente) no deben arrastrar el base64 del PDF. Manteniéndolo en una colección
aparte, esas queries siguen livianas.

### Envío de email con adjunto

`EmailService` hoy arma un payload Resend con `from/to/subject/html` y hace POST a
`https://api.resend.com/emails`. Resend soporta `attachments: [{ filename, content }]`
donde `content` es el archivo en base64. Se agrega soporte de adjuntos:

- Nuevo método `SendInvoice(order *domain.Order, pdfBase64, filename string) error`
  que arma el payload Resend con `attachments: [{ filename, content: pdfBase64 }]`.
  El `Send` existente no se toca (se puede factorizar un helper interno que acepte
  adjuntos opcionales, sin cambiar la firma pública de `Send`).
- **Destinatario: SOLO el cliente** (`order.CustomerEmail`). El mail de factura NO
  se envía a `OwnerEmails` — a diferencia de los avisos de nuevo pedido / cambio de
  estado / registro, este flujo no notifica a los dueños.
- Asunto: `Factura de tu pedido #XXXX — Pooled`.
- Cuerpo HTML fijo y simple, con el número de pedido; el PDF va adjunto.

### Endpoints (admin, detrás de auth + admin middleware)

- `POST /admin/orders/:id/invoice` — multipart con el archivo `file`. Valida que sea
  PDF y el tamaño; guarda/reemplaza el doc en `invoices` (base64); setea
  `Order.InvoiceSentAt`; envía el email con el PDF adjunto. Devuelve la fecha de envío.
- `POST /admin/orders/:id/invoice/resend` — reenvía la factura ya guardada (lee el
  base64 de `invoices`, remanda el mail, actualiza `sent_at`). Falla con 404 si el
  pedido no tiene factura.

### Frontend (admin)

En `OrderDetailDialog` (detalle del pedido), nueva sección **"Factura"**:
- Si el pedido NO tiene factura: input de archivo (solo `.pdf`) + botón
  "Enviar factura al cliente".
- Si el pedido YA tiene factura: muestra "Factura enviada el {fecha}" + botón
  "Reenviar" + opción de subir una nueva (reemplaza).
- En la fila de la tabla de pedidos (`OrdersTable`), un indicador "Factura enviada"
  cuando `invoice_sent_at` está presente.

Sube vía `multipart/form-data` a los endpoints admin. React Query invalida el
listado de pedidos al enviar para reflejar el estado.

## Flujo de datos

1. Admin abre el detalle del pedido → sección Factura.
2. Elige un PDF → `POST /admin/orders/:id/invoice` (multipart).
3. Backend valida (PDF, ≤8 MB) → guarda base64 en `invoices` (upsert por `order_id`)
   → setea `Order.InvoiceSentAt` → envía email con adjunto al `CustomerEmail`.
4. Respuesta con `sent_at` → el frontend muestra "Factura enviada el {fecha}".
5. Reenvío: `POST /admin/orders/:id/invoice/resend` → lee base64 → remanda.

## Manejo de errores

- Archivo no PDF o > 8 MB → 400 con mensaje claro; no se guarda ni envía.
- Falla el envío de email (Resend) → la factura igual queda guardada; se devuelve
  error indicando que se guardó pero no se pudo enviar, para poder reintentar con
  "Reenviar". (El guardado y el envío son pasos separados: no perder el archivo si
  el mail falla.)
- Reenvío sin factura previa → 404.

## Fuera de scope (YAGNI)

- Descarga de la factura por el cliente desde el sitio.
- Texto del mail editable.
- Múltiples facturas / notas de crédito por pedido.
- Generación automática de la factura (el admin la sube ya hecha).

## Riesgos / notas

- Límite de 16 MB por documento Mongo: cubierto con el tope de subida de 8 MB.
- base64 infla ~33% el tamaño en la base; aceptable para PDFs de factura chicos.
- El `content` base64 nunca se incluye en respuestas de listado de pedidos, solo se
  lee al reenviar y al construir el adjunto.
