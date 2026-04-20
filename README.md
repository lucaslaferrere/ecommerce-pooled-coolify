# Ecommerce Backend API

Backend de e-commerce construido en **Go** siguiendo la arquitectura **Clean Architecture / Hexagonal**.

## 🚀 Quick Start

### Requisitos Previos
- Go 1.25.6 o superior
- MongoDB 5.0 o superior
- Git

### Instalación

1. Clonar el repositorio
```bash
git clone <repository-url>
cd ecommerce-pooled
```

2. Descargar dependencias
```bash
go mod download
go mod tidy
```

3. Configurar variables de entorno
```bash
# Crear archivo .env
PORT=8080
MONGO_URI=mongodb://localhost:27017
DB_NAME=ecommerce
GIN_MODE=debug
JWT_SECRET=your-secret-key-here
```

4. Iniciar MongoDB
```bash
# Usando Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# O usar MongoDB localmente
mongod
```

5. Ejecutar la aplicación
```bash
go run cmd/api/main.go
```

## 📁 Estructura del Proyecto

```
ecommerce-pooled/
├── cmd/
│   └── api/                    # Punto de entrada
│       └── main.go
├── internal/
│   ├── adapters/              # Implementaciones concretas
│   │   ├── database/          # Conexión MongoDB
│   │   ├── handlers/          # HTTP handlers (Gin)
│   │   └── repositories/      # Persistencia (Mongo)
│   ├── config/                # Configuración
│   ├── core/                  # Lógica de negocio
│   │   ├── domain/            # Entidades
│   │   ├── ports/             # Interfaces/Contratos
│   │   └── services/          # Casos de uso
│   └── utils/                 # Utilidades
├── go.mod
├── go.sum
├── ARCHITECTURE.md
└── README.md
```

## 🏗️ Arquitectura

Este proyecto implementa **Clean Architecture** (también conocida como **Hexagonal Architecture**):

### Capas

1. **Domain (Núcleo)**: Contiene la lógica de negocio pura
   - Entidades sin dependencias externas
   - Reglas de negocio

2. **Ports**: Define interfaces/contratos
   - Repositories: Contratos de persistencia
   - Services: Contratos de servicios

3. **Adapters**: Implementaciones concretas
   - HTTP Handlers: Controladores Gin
   - Repositories: Implementaciones MongoDB
   - Database: Configuración de conexión

### Ventajas de esta Arquitectura

- ✅ Independencia de frameworks
- ✅ Fácil testing
- ✅ Escalabilidad
- ✅ Mantenibilidad
- ✅ Separación de responsabilidades

## 📦 Entidades Principales

### User
```go
type User struct {
    ID           ObjectID
    Email        string
    PasswordHash string
    Role         string // "admin" | "client"
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Product
```go
type Product struct {
    ID          ObjectID
    Name        string
    Description string
    BasePrice   float64
    Category    string
    Brand       string
    Images      []string
    Variants    []Variant
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Variant struct {
    SKU             string
    Color           string
    Size            string
    Stock           int
    PriceAdjustment float64
}
```

### Order
```go
type Order struct {
    ID        ObjectID
    UserID    ObjectID
    Items     []CartItem
    Total     float64
    Status    string // "pending", "processing", "shipped", "delivered", "cancelled"
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### CartItem
```go
type CartItem struct {
    ProductID  ObjectID
    VariantSKU string
    Quantity   int
}
```

## 🔌 Interfaces (Ports)

### UserRepository
- `Create(ctx context.Context, user *User) error`
- `GetByID(ctx context.Context, id ObjectID) (*User, error)`
- `GetByEmail(ctx context.Context, email string) (*User, error)`
- `Update(ctx context.Context, user *User) error`
- `Delete(ctx context.Context, id ObjectID) error`
- `List(ctx context.Context, skip int64, limit int64) ([]*User, error)`

### ProductRepository
- `Create(ctx context.Context, product *Product) error`
- `GetByID(ctx context.Context, id ObjectID) (*Product, error)`
- `Update(ctx context.Context, product *Product) error`
- `Delete(ctx context.Context, id ObjectID) error`
- `List(ctx context.Context, filter map[string]interface{}, skip int64, limit int64) ([]*Product, error)`
- `GetByCategory(ctx context.Context, category string, skip int64, limit int64) ([]*Product, error)`
- `GetByBrand(ctx context.Context, brand string, skip int64, limit int64) ([]*Product, error)`
- `UpdateVariantStock(ctx context.Context, productID ObjectID, sku string, newStock int) error`

### OrderRepository
- `Create(ctx context.Context, order *Order) error`
- `GetByID(ctx context.Context, id ObjectID) (*Order, error)`
- `GetByUserID(ctx context.Context, userID ObjectID, skip int64, limit int64) ([]*Order, error)`
- `Update(ctx context.Context, order *Order) error`
- `UpdateStatus(ctx context.Context, id ObjectID, status string) error`
- `Delete(ctx context.Context, id ObjectID) error`
- `List(ctx context.Context, skip int64, limit int64) ([]*Order, error)`
- `GetByStatus(ctx context.Context, status string, skip int64, limit int64) ([]*Order, error)`

## 🔧 Tecnologías Utilizadas

- **Go**: 1.25.6
- **Gin**: Framework HTTP de alto rendimiento
- **MongoDB**: Base de datos NoSQL
- **mongo-go-driver**: Driver oficial de MongoDB para Go

## 📝 Próximos Pasos

- [ ] Implementar lógica de servicios (core/services)
- [ ] Implementar handlers HTTP (adapters/handlers)
- [ ] Crear middleware de autenticación JWT
- [ ] Añadir validaciones
- [ ] Implementar paginación
- [ ] Añadir tests unitarios
- [ ] Documentación OpenAPI/Swagger
- [ ] Dockerización

## 📄 Licencia

MIT

## 👤 Autor

Senior Software Engineer - Clean Architecture Specialist
