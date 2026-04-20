# API Endpoints - Ecommerce Backend

## Base URL
```
http://localhost:8080/api
```

## Health Check

### GET /health
```bash
curl http://localhost:8080/health
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

## Users API

### Create User
**POST** `/users`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "role": "client"
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "role": "client",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Get User by ID
**GET** `/users/:id`

**Example:**
```bash
curl http://localhost:8080/api/users/507f1f77bcf86cd799439011
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "role": "client",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Get User by Email
**GET** `/users/email/:email`

**Example:**
```bash
curl http://localhost:8080/api/users/email/user@example.com
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "role": "client",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Update User
**PUT** `/users/:id`

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "role": "client"
}
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "newemail@example.com",
  "role": "client",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Delete User
**DELETE** `/users/:id`

**Response (200 OK):**
```json
{
  "message": "usuario eliminado"
}
```

---

## Products API

### Create Product
**POST** `/products`

**Request Body:**
```json
{
  "name": "Laptop Dell XPS 13",
  "description": "Laptop ultraligera de última generación",
  "base_price": 1299.99,
  "category": "Electrónica",
  "brand": "Dell",
  "images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "variants": [
    {
      "sku": "DELL-XPS-13-SLV-8GB",
      "color": "Silver",
      "size": "13.3\"",
      "stock": 50,
      "price_adjustment": 0
    },
    {
      "sku": "DELL-XPS-13-BLK-16GB",
      "color": "Black",
      "size": "13.3\"",
      "stock": 30,
      "price_adjustment": 200
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "Laptop Dell XPS 13",
  "description": "Laptop ultraligera de última generación",
  "base_price": 1299.99,
  "category": "Electrónica",
  "brand": "Dell",
  "images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "variants": [
    {
      "sku": "DELL-XPS-13-SLV-8GB",
      "color": "Silver",
      "size": "13.3\"",
      "stock": 50,
      "price_adjustment": 0
    },
    {
      "sku": "DELL-XPS-13-BLK-16GB",
      "color": "Black",
      "size": "13.3\"",
      "stock": 30,
      "price_adjustment": 200
    }
  ],
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Get Product
**GET** `/products/:id`

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "Laptop Dell XPS 13",
  "description": "Laptop ultraligera de última generación",
  "base_price": 1299.99,
  "category": "Electrónica",
  "brand": "Dell",
  "images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "variants": [
    {
      "sku": "DELL-XPS-13-SLV-8GB",
      "color": "Silver",
      "size": "13.3\"",
      "stock": 50,
      "price_adjustment": 0
    }
  ],
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### List Products by Category
**GET** `/products/category/:category?skip=0&limit=10`

**Example:**
```bash
curl "http://localhost:8080/api/products/category/Electrónica?skip=0&limit=10"
```

**Response (200 OK):**
```json
{
  "products": [
    {
      "id": "507f1f77bcf86cd799439012",
      "name": "Laptop Dell XPS 13",
      "base_price": 1299.99,
      "category": "Electrónica",
      "brand": "Dell"
    }
  ],
  "total": 1
}
```

### Update Product
**PUT** `/products/:id`

**Request Body:**
```json
{
  "name": "Laptop Dell XPS 13 Updated",
  "base_price": 1399.99
}
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "Laptop Dell XPS 13 Updated",
  "base_price": 1399.99
}
```

### Delete Product
**DELETE** `/products/:id`

**Response (200 OK):**
```json
{
  "message": "producto eliminado"
}
```

---

## Orders API

### Create Order
**POST** `/orders`

**Request Body:**
```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "items": [
    {
      "product_id": "507f1f77bcf86cd799439012",
      "variant_sku": "DELL-XPS-13-SLV-8GB",
      "quantity": 1
    }
  ],
  "total": 1299.99,
  "status": "pending"
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439013",
  "user_id": "507f1f77bcf86cd799439011",
  "items": [
    {
      "product_id": "507f1f77bcf86cd799439012",
      "variant_sku": "DELL-XPS-13-SLV-8GB",
      "quantity": 1
    }
  ],
  "total": 1299.99,
  "status": "pending",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Get Order
**GET** `/orders/:id`

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439013",
  "user_id": "507f1f77bcf86cd799439011",
  "items": [
    {
      "product_id": "507f1f77bcf86cd799439012",
      "variant_sku": "DELL-XPS-13-SLV-8GB",
      "quantity": 1
    }
  ],
  "total": 1299.99,
  "status": "pending",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:30:00Z"
}
```

### Get User Orders
**GET** `/users/:userID/orders?skip=0&limit=10`

**Example:**
```bash
curl "http://localhost:8080/api/users/507f1f77bcf86cd799439011/orders?skip=0&limit=10"
```

**Response (200 OK):**
```json
{
  "orders": [
    {
      "id": "507f1f77bcf86cd799439013",
      "user_id": "507f1f77bcf86cd799439011",
      "items": [],
      "total": 1299.99,
      "status": "pending"
    }
  ],
  "total": 1
}
```

### Update Order Status
**PUT** `/orders/:id/status`

**Request Body:**
```json
{
  "status": "processing"
}
```

**Valid Statuses:**
- `pending` - Orden pendiente
- `processing` - En procesamiento
- `shipped` - Enviada
- `delivered` - Entregada
- `cancelled` - Cancelada

**Response (200 OK):**
```json
{
  "message": "estado actualizado"
}
```

### Cancel Order
**PUT** `/orders/:id/cancel`

**Response (200 OK):**
```json
{
  "message": "orden cancelada"
}
```

### Delete Order
**DELETE** `/orders/:id`

**Response (200 OK):**
```json
{
  "message": "orden eliminada"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid request parameters"
}
```

### 404 Not Found
```json
{
  "error": "usuario no encontrado"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Common Query Parameters

| Parámetro | Tipo | Descripción | Por Defecto |
|-----------|------|-------------|------------|
| `skip` | integer | Número de registros a saltar (paginación) | 0 |
| `limit` | integer | Número máximo de registros a retornar | 10 |

---

## Rate Limiting & Authentication

> **Nota:** La autenticación y rate limiting serán implementados en futuras iteraciones.

---

## Testing con cURL

### Crear usuario
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "role": "client"
  }'
```

### Obtener usuario
```bash
curl http://localhost:8080/api/users/507f1f77bcf86cd799439011
```

### Crear producto
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Producto Test",
    "description": "Descripción",
    "base_price": 99.99,
    "category": "Electrónica",
    "brand": "TestBrand",
    "images": [],
    "variants": [
      {
        "sku": "TEST-SKU-001",
        "color": "Red",
        "size": "M",
        "stock": 100,
        "price_adjustment": 0
      }
    ]
  }'
```
