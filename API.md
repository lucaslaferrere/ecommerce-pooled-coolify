# Pooled API — Documentación

**Base URL:** `https://api.pooled.com.ar/api/v1`

**Autenticación:** Bearer Token en el header `Authorization: Bearer <token>`

---

## Autenticación

### `POST /auth/register`
Registra un nuevo usuario. El rol siempre es `client` — no se puede cambiar desde el cliente.

```json
// Body
{
  "email": "user@example.com",
  "password": "min6chars"
}

// Response 201
{
  "id": "...",
  "email": "user@example.com",
  "role": "client",
  "created_at": "2026-06-19T13:00:00Z"
}
```

---

### `POST /auth/login`

```json
// Body
{
  "email": "user@example.com",
  "password": "..."
}

// Response 200
{
  "access_token": "<JWT>",
  "refresh_token": "<token>"
}
```

---

### `POST /auth/refresh`

```json
// Body
{ "refresh_token": "..." }

// Response 200
{
  "access_token": "<JWT>",
  "refresh_token": "<token>"
}
```

---

### `POST /auth/forgot-password`

```json
// Body
{ "email": "user@example.com" }

// Response 200 — siempre OK (no revela si el email existe)
```

---

### `POST /auth/reset-password`

```json
// Body
{
  "email": "user@example.com",
  "code": "123456",
  "new_password": "min6chars"
}
```

---

## Productos

> Endpoints públicos — no requieren autenticación.

### `GET /products`
Lista productos visibles.

**Query params:**

| Param      | Tipo   | Default | Descripción             |
|------------|--------|---------|-------------------------|
| `skip`     | int    | 0       | Offset de paginación    |
| `limit`    | int    | 20      | Resultados por página   |
| `category` | string | —       | Filtrar por categoría   |
| `brand`    | string | —       | Filtrar por marca       |

```json
// Response 200
[
  {
    "id": "64f1a2b3c4d5e6f7a8b9c0d1",
    "name": "HORUS 18W WW",
    "description": "...",
    "category": "iluminacion",
    "brand": "Poolside",
    "base_price": 180987.52,
    "discount_percent": 0,
    "stock": 10,
    "images": ["/uploads/horus-18w.jpg"],
    "variants": [],
    "specs": {},
    "main_specs": [],
    "visible": true,
    "created_at": "2026-06-01T00:00:00Z"
  }
]
```

**Objeto Variant:**
```json
{
  "sku": "AZUL-M",
  "color": "Azul",
  "size": "M",
  "stock": 5,
  "price_adjustment": 0.0
}
```

---

### `GET /products/:id`
Producto por ID.

---

### `GET /products/category/:category`
Lista productos por categoría.

---

### `GET /products/brand/:brand`
Lista productos por marca.

---

## Kits

> Endpoints públicos — no requieren autenticación.

### `GET /kits`

**Query params:** `limit`, `featured` (boolean)

```json
// Response 200
[
  {
    "id": "...",
    "name": "Kit Alberca Premium",
    "price": 670541.99,
    "image_url": "/uploads/kit-alberca.jpg",
    "visible": true
  }
]
```

---

### `GET /kits/:id`

---

## Órdenes

> Requieren `Authorization: Bearer <access_token>`

### `POST /checkout`
Crea una orden. El backend calcula precio, descuento y costo de envío — estos valores **no se aceptan del cliente**.

```json
// Body
{
  "customer_name": "Juan Pérez",
  "customer_email": "juan@example.com",
  "customer_phone": "1123456789",
  "dni_cuit": "20-12345678-9",
  "items": [
    {
      "product_id": "64f1a2b3c4d5e6f7a8b9c0d1",
      "variant_sku": "AZUL-M",
      "quantity": 2
    }
  ],
  "payment_method": "transferencia",
  "delivery_method": "envio",
  "shipping_address": "Av. Siempre Viva 123",
  "shipping_city": "Buenos Aires",
  "shipping_province": "CABA",
  "shipping_zip": "1425",
  "notes": "Sin timbre",
  "factura_a": {
    "razon_social": "Mi Empresa S.A.",
    "cuit": "30-12345678-9"
  }
}
```

**Valores válidos:**

| Campo             | Opciones                          |
|-------------------|-----------------------------------|
| `payment_method`  | `transferencia`, `mercadopago`    |
| `delivery_method` | `envio`, `retirar`                |

```json
// Response 201
{
  "order": {
    "id": "...",
    "status": "pending",
    "total": 361134.74,
    "shipping_cost": 9000.0,
    "discount": 0.0
  },
  "init_point": "https://www.mercadopago.com.ar/...",
  "sandbox_init_point": "https://sandbox.mercadopago.com.ar/..."
}
```

> `init_point` y `sandbox_init_point` solo se incluyen si `payment_method` es `mercadopago`.

---

### `GET /orders/me`
Órdenes del usuario autenticado.

**Query params:** `skip`, `limit`

---

### `GET /orders/:id`
Orden por ID. Solo accesible por el dueño de la orden.

---

### `PATCH /orders/:id/cancel`
Cancela una orden propia. Solo funciona en estado `pending` o `processing`.

---

## Usuarios

> El usuario autenticado solo puede acceder/modificar su propio perfil. Los admins pueden acceder a cualquier usuario.

### `GET /users/:id`

```json
// Response 200
{
  "id": "...",
  "email": "user@example.com",
  "role": "client",
  "created_at": "..."
}
```

---

### `PUT /users/:id`
Actualiza el email del usuario. El rol no es modificable por este endpoint.

```json
// Body
{ "email": "nuevo@example.com" }
```

---

## Admin — Órdenes

> Requieren token con `role: "admin"`

### `GET /admin/orders`

**Query params:**

| Param    | Tipo   | Default | Descripción                                                                 |
|----------|--------|---------|-----------------------------------------------------------------------------|
| `page`   | int    | 1       | Página                                                                      |
| `limit`  | int    | 20      | Resultados por página                                                       |
| `status` | string | —       | `pending` \| `paid` \| `processing` \| `shipped` \| `delivered` \| `cancelled` |

```json
// Response 200
{
  "orders": [ { ...Order } ],
  "total": 142,
  "page": 1,
  "limit": 20
}
```

**Objeto Order completo:**
```json
{
  "id": "string",
  "user_id": "string",
  "customer_name": "string",
  "customer_email": "string",
  "customer_phone": "string",
  "dni_cuit": "string",
  "status": "pending | paid | processing | shipped | delivered | cancelled | rejected",
  "total": 361134.74,
  "shipping_cost": 9000.0,
  "discount": 12638.57,
  "payment_method": "transferencia | mercadopago",
  "delivery_method": "envio | retirar",
  "shipping_details": {
    "address": "Av. Siempre Viva 123",
    "city": "Buenos Aires",
    "province": "CABA",
    "postal_code": "1425"
  },
  "items": [
    {
      "product_id": "string",
      "name": "string",
      "variant_sku": "string",
      "quantity": 2,
      "unit_price": 180987.52,
      "item_type": "product | kit"
    }
  ],
  "factura_a": {
    "razon_social": "Mi Empresa S.A.",
    "cuit": "30-12345678-9"
  },
  "tracking_number": "AND123456",
  "notes": "string",
  "preference_id": "string",
  "payment_id": "string",
  "created_at": "2026-06-19T13:00:00Z",
  "updated_at": "2026-06-19T13:00:00Z"
}
```

---

### `PATCH /admin/orders/:id/status`

```json
// Body
{
  "status": "shipped",
  "tracking_number": "AND123456"
}
```

> `tracking_number` es opcional. Se recomienda incluirlo cuando `status` es `shipped`.

---

### `DELETE /admin/orders/:id`

---

## Admin — Usuarios

### `GET /admin/users`

**Query params:** `skip`, `limit`

```json
// Response 200
{
  "users": [ { ...User } ],
  "total": 38
}
```

---

### `DELETE /admin/users/:id`

---

## Admin — Productos

### `GET /admin/products`
Lista todos los productos (incluye los no visibles).

### `POST /admin/products`
Crea un producto. Requiere `multipart/form-data`.

| Campo        | Tipo        | Requerido | Descripción                        |
|--------------|-------------|-----------|-------------------------------------|
| `name`       | string      | ✓         |                                     |
| `category`   | string      | ✓         |                                     |
| `base_price` | string (número) | ✓    | Ej: `"180987.52"`                  |
| `description`| string      |           |                                     |
| `brand`      | string      |           |                                     |
| `stock`      | string (int)|           | Stock general (sin variantes)       |
| `variants`   | JSON string |           | Array de objetos Variant            |
| `specs`      | JSON string |           | Objeto clave-valor                  |
| `main_specs` | JSON string |           | Array de specs destacadas           |
| `image`      | file        |           | Imagen principal (multipart)        |

### `PUT /admin/products/:id`
Actualiza un producto. Mismos campos que POST.

### `PATCH /admin/products/:id/visibility`
```json
// Body
{ "visible": true }
```

### `PATCH /admin/products/:id/sort-order`
```json
// Body
{ "sort_order": 3 }
```

### `PATCH /admin/products/:id/variants/:sku/stock`
```json
// Body
{ "stock": 10 }
```

### `DELETE /admin/products/:id`

---

## Admin — Kits

### `GET /admin/kits`
Lista todos los kits (incluye los no visibles).

### `POST /admin/kits`
### `PUT /admin/kits/:id`
### `PATCH /admin/kits/:id/visibility`
### `PATCH /admin/kits/:id/sort-order`
### `DELETE /admin/kits/:id`

---

## Admin — Cart Links

Links pre-armados que cargan un carrito al abrirlos. Útil para ventas asistidas.

### `POST /admin/cart-links`

```json
// Body
{
  "items": [
    {
      "product_id": "64f1a2b3c4d5e6f7a8b9c0d1",
      "variant_sku": "",
      "name": "HORUS 18W WW",
      "image_url": "https://api.pooled.com.ar/uploads/horus.jpg",
      "unit_price": 180987.52,
      "quantity": 2,
      "type": "product"
    }
  ]
}
```

> `type` acepta `"product"` o `"kit"`. `variant_sku` puede ser vacío para productos sin variantes.

```json
// Response 201
{
  "token": "abc123def456...",
  "url": "https://pooled.com.ar/checkout?cart=abc123def456...",
  "expires_at": "2026-07-19T13:00:00Z"
}
```

### `GET /cart-links/:token` (público)
Retorna los items del carrito para ese token. Expira a los 30 días.

---

## Admin — Analytics

### `GET /admin/analytics`
Eventos agrupados por tipo.

### `GET /admin/traffic`
Tráfico por día/hora.

---

## Formularios (público)

### `POST /distributor-leads`

```json
{
  "name": "string",
  "email": "string",
  "phone": "string",
  "company": "string",
  "message": "string"
}
```

### `POST /warranty`
Solicitud de garantía de producto.

### `POST /wizard-recommendations`
Resultado del wizard de recomendación de producto.

### `POST /events`
Evento de analytics (fire-and-forget, sin auth requerida).

---

## Webhooks

### `POST /webhooks/mercadopago`
Uso exclusivo de MercadoPago. Valida firma HMAC-SHA256 si la variable `MP_WEBHOOK_SECRET` está configurada en el servidor.

---

## Códigos de error comunes

| Código | Significado                                      |
|--------|--------------------------------------------------|
| 400    | Body inválido o campos faltantes                 |
| 401    | Token ausente, inválido o expirado               |
| 403    | Sin permisos (ej: intentar acceder a otro usuario) |
| 404    | Recurso no encontrado                            |
| 409    | Conflicto (ej: email ya registrado)              |
| 422    | Stock insuficiente u otro error de validación de negocio |
| 500    | Error interno del servidor                       |

```json
// Formato de error
{ "error": "descripción del error" }
```

---

## Notificaciones de email

### Al cliente
| Evento | Asunto |
|--------|--------|
| Checkout completado | `✅ Pedido confirmado #XXXXXXXX — Pooled` |
| Admin cambia estado a `paid` | `✅ Pedido confirmado #XXXXXXXX — Pooled` |
| Admin cambia estado a `shipped` | `🚚 Tu pedido #XXXXXXXX está en camino — Pooled` (incluye número de tracking) |

### A los dueños (notificación interna)
| Evento | Asunto |
|--------|--------|
| Nuevo pedido (cualquier método de pago) | `🛒 Nuevo pedido #XXXXXXXX — Nombre Cliente` |
| Admin cambia cualquier estado | `📋 Pedido #XXXXXXXX → [Estado] — Nombre Cliente` |

Los emails a dueños se envían a todos los destinatarios configurados en la variable de entorno `OWNER_EMAILS` (lista separada por comas).

**Variable de entorno requerida:**
```
OWNER_EMAILS=admin@ejemplo.com,otro@ejemplo.com
```

---

## Notas para integración con CRM

- Usar `GET /admin/orders?status=paid` para obtener pedidos confirmados listos para procesar.
- El campo `customer_email` identifica al cliente en todos los pedidos.
- `PATCH /admin/orders/:id/status` con `status: "processing"` para marcar como en preparación, `"shipped"` con `tracking_number` para notificar envío.
- Los pedidos en estado `pending` con más de 24hs sin pago son candidatos a seguimiento.
- El campo `payment_method: "transferencia"` requiere confirmación manual — verificar comprobante antes de cambiar a `paid`.
