# 🔐 API Reference - Autenticación y Autorización

## 📋 Tabla Rápida

| Método | Endpoint | Autenticación | Autorización | Descripción |
|--------|----------|---------------|--------------|-------------|
| POST | `/auth/register` | No | - | Registrar usuario |
| POST | `/auth/login` | No | - | Iniciar sesión |
| POST | `/auth/refresh` | No | - | Refrescar token |
| GET | `/api/users/:id` | No | - | Obtener usuario público |
| GET | `/api/users/email/:email` | No | - | Obtener por email público |
| GET | `/api/users` | Sí | - | Listar usuarios |
| PUT | `/api/users/:id` | Sí | - | Actualizar usuario |
| DELETE | `/api/users/:id` | Sí | - | Eliminar usuario |
| DELETE | `/api/users/admin/:id` | Sí | Admin | Eliminar usuario (admin) |
| POST | `/api/products` | Sí | Admin | Crear producto |
| PUT | `/api/products/:id` | Sí | Admin | Actualizar producto |
| DELETE | `/api/products/:id` | Sí | Admin | Eliminar producto |
| POST | `/api/orders` | Sí | - | Crear orden |
| GET | `/api/orders/admin/list` | Sí | Admin | Listar todas las órdenes |

---

## 🔓 Endpoints Públicos (Sin Autenticación)

### POST /auth/register
Registra un nuevo usuario en el sistema.

**Request:**
```json
{
  "email": "usuario@example.com",
  "password": "micontraseña123",
  "role": "client"
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "usuario@example.com",
  "role": "client",
  "created_at": "2026-04-16T10:30:00Z"
}
```

**Errores:**
- 400: Email ya registrado
- 400: Email inválido
- 400: Contraseña muy corta (< 6 caracteres)
- 400: Role inválido

---

### POST /auth/login
Inicia sesión y retorna tokens JWT.

**Request:**
```json
{
  "email": "usuario@example.com",
  "password": "micontraseña123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

**Errores:**
- 401: Credenciales inválidas
- 400: Email o contraseña requeridos

---

### POST /auth/refresh
Genera un nuevo access token usando el refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

**Errores:**
- 401: Refresh token inválido
- 401: Tipo de token inválido

---

## 🔒 Endpoints Protegidos (Requieren Autenticación)

### GET /api/users
Obtiene lista de usuarios con paginación.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `skip` (int, default: 0) - Registros a saltar
- `limit` (int, default: 10) - Máximo de registros

**Request:**
```bash
curl -X GET "http://localhost:8080/api/users?skip=0&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "507f1f77bcf86cd799439011",
      "email": "usuario1@example.com",
      "role": "client",
      "created_at": "2026-04-16T10:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439012",
      "email": "usuario2@example.com",
      "role": "admin",
      "created_at": "2026-04-16T10:35:00Z"
    }
  ],
  "total": 2
}
```

**Errores:**
- 401: Token requerido
- 401: Token inválido
- 500: Error en servidor

---

### PUT /api/users/:id
Actualiza información del usuario.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
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
  "updated_at": "2026-04-16T10:40:00Z"
}
```

**Errores:**
- 401: Autenticación requerida
- 400: ID inválido
- 500: Error en servidor

---

### DELETE /api/users/:id
Elimina un usuario.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "message": "usuario eliminado"
}
```

**Errores:**
- 401: Autenticación requerida
- 400: ID inválido
- 500: Usuario no encontrado

---

## 👨‍💼 Endpoints Solo Admin

### DELETE /api/users/admin/:id
Elimina un usuario (solo administrador).

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Response (200 OK):**
```json
{
  "message": "usuario eliminado"
}
```

**Errores:**
- 401: Autenticación requerida
- 403: Se requiere rol admin
- 400: ID inválido

---

### POST /api/products
Crea un nuevo producto (solo administrador).

**Headers:**
```
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Laptop Dell XPS",
  "description": "Laptop ultraligera",
  "base_price": 1299.99,
  "category": "Electrónica",
  "brand": "Dell",
  "images": ["https://example.com/img1.jpg"],
  "variants": [
    {
      "sku": "XPS-001",
      "color": "Silver",
      "size": "13.3",
      "stock": 50,
      "price_adjustment": 0
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439020",
  "name": "Laptop Dell XPS",
  "base_price": 1299.99,
  "category": "Electrónica",
  "created_at": "2026-04-16T10:50:00Z"
}
```

**Errores:**
- 401: Autenticación requerida
- 403: Se requiere rol admin
- 400: Datos inválidos

---

### GET /api/orders/admin/list
Lista todas las órdenes del sistema (solo administrador).

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Query Parameters:**
- `skip` (int, default: 0)
- `limit` (int, default: 10)

**Response (200 OK):**
```json
{
  "orders": [
    {
      "id": "507f1f77bcf86cd799439030",
      "user_id": "507f1f77bcf86cd799439011",
      "total": 1299.99,
      "status": "pending",
      "created_at": "2026-04-16T11:00:00Z"
    }
  ],
  "total": 1
}
```

**Errores:**
- 401: Autenticación requerida
- 403: Se requiere rol admin

---

## 🔑 Cómo Usar Tokens

### 1. Obtener Token
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "password": "contraseña"
  }'

# Guardar el access_token
```

### 2. Usar Token en Peticiones
```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer <tu_access_token>"
```

### 3. Refrescar Token cuando Expira
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<tu_refresh_token>"
  }'
```

---

## ⏱️ Tiempos de Expiración

| Token | Expiración | Uso |
|-------|-----------|-----|
| Access Token | 15 minutos | Acceder a recursos protegidos |
| Refresh Token | 7 días | Renovar access token |

---

## 🛡️ Estructura de Header de Autenticación

```
Authorization: Bearer <access_token>
```

**Formato correcto:**
- Debe incluir la palabra "Bearer"
- Seguida de un espacio
- Seguida del token JWT

**Ejemplos:**
```
✅ Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
❌ Authorization: eyJhbGciOiJIUzI1NiIs...
❌ Authorization: JWT eyJhbGciOiJIUzI1NiIs...
```

---

## 🔍 Contenido del JWT (Claims)

Cuando decodificas el JWT, contiene:

```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "email": "usuario@example.com",
  "role": "admin",
  "type": "access",
  "exp": 1713270000,
  "iat": 1713269100
}
```

---

## 📝 Ejemplos Completos

### Flujo Completo de Autenticación

```bash
# 1. Registrarse
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepass123",
    "role": "client"
  }'

# 2. Iniciar sesión
RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepass123"
  }')

TOKEN=$(echo $RESPONSE | jq -r '.access_token')

# 3. Usar token para acceder a recursos protegidos
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN"

# 4. Cuando expire, refrescar token
REFRESH=$(echo $RESPONSE | jq -r '.refresh_token')

curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH\"}"
```

---

## ⚠️ Códigos de Error Comunes

| Código | Significado | Solución |
|--------|------------|----------|
| 400 | Bad Request | Verifica los datos enviados |
| 401 | Unauthorized | Incluye token válido en Authorization |
| 403 | Forbidden | El rol del usuario no tiene permiso |
| 404 | Not Found | El recurso no existe |
| 500 | Server Error | Error interno del servidor |

---

## 🔒 Buenas Prácticas

✅ **Almacenar tokens de forma segura** (localStorage, sessionStorage)  
✅ **Usar HTTPS en producción** (no HTTP)  
✅ **Refrescar token antes de expirar**  
✅ **No exponer tokens en logs**  
✅ **Cambiar JWT_SECRET en producción**  
✅ **Usar tokens diferentes para acceso y refresco**  
✅ **Implementar logout (eliminar tokens client-side)**  
✅ **Validar email cuando sea necesario**  

---

*API Reference - Autenticación y Autorización*  
*Última actualización: 16 de Abril de 2026*
