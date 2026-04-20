# 🛠️ Guía de Desarrollo - Ecommerce Backend

## Requisitos de Desarrollo

- Go 1.25.6 o superior
- MongoDB 5.0 o superior
- Git
- Un editor de código (GoLand, VS Code, etc.)

## Setup Inicial

### 1. Clonar el Repositorio
```bash
git clone <repository-url>
cd ecommerce-pooled
```

### 2. Descargar Dependencias
```bash
go mod download
go mod tidy
```

### 3. Configurar Variables de Entorno
```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar .env con tus valores locales
# Por defecto, apunta a MongoDB en localhost:27017
```

### 4. Iniciar MongoDB

#### Opción 1: Con Docker (Recomendado)
```bash
# Crear contenedor MongoDB
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:latest

# Ver logs
docker logs mongodb

# Conectar a MongoDB desde shell
docker exec -it mongodb mongosh -u admin -p password
```

#### Opción 2: MongoDB Local
```bash
# En Windows
# 1. Descargar desde: https://www.mongodb.com/try/download/community
# 2. Instalar
# 3. Ejecutar en PowerShell:
mongod

# En macOS
brew services start mongodb-community

# En Linux
sudo systemctl start mongod
```

### 5. Ejecutar la Aplicación
```bash
go run cmd/api/main.go
```

El servidor debería iniciar en `http://localhost:8080`

## Verificar que Todo Funciona

### Health Check
```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

## Estructura de Desarrollo

### Arquitectura por Capas

```
Tier 1: HTTP Layer (Handlers)
├── /adapters/handlers/user_handler.go
├── /adapters/handlers/product_handler.go
└── /adapters/handlers/order_handler.go

Tier 2: Service Layer (Lógica de Negocio)
├── /core/services/user_service.go
├── /core/services/product_service.go
└── /core/services/order_service.go

Tier 3: Port Layer (Interfaces)
├── /core/ports/user_repository.go
├── /core/ports/product_repository.go
└── /core/ports/order_repository.go

Tier 4: Domain Layer (Modelos Puros)
├── /core/domain/user.go
├── /core/domain/product.go
├── /core/domain/cart_item.go
└── /core/domain/order.go

Tier 5: Adapter Layer (Implementaciones)
├── /adapters/repositories/user_repository_mongo.go
├── /adapters/repositories/product_repository_mongo.go
├── /adapters/repositories/order_repository_mongo.go
└── /adapters/database/mongo.go
```

## Flujo de Desarrollo

### Agregar una Nueva Entidad

1. **Crear el struct de dominio** en `/internal/core/domain/`
```go
// ejemplo_entity.go
package domain

type Ejemplo struct {
    ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name string             `bson:"name" json:"name"`
    // ... más campos
}
```

2. **Definir la interface** en `/internal/core/ports/`
```go
// ejemplo_repository.go
package ports

type EjemploRepository interface {
    Create(ctx context.Context, ejemplo *domain.Ejemplo) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Ejemplo, error)
    // ... más métodos
}
```

3. **Implementar el servicio** en `/internal/core/services/`
```go
// ejemplo_service.go
package services

type EjemploService struct {
    repo ports.EjemploRepository
}

func NewEjemploService(repo ports.EjemploRepository) *EjemploService {
    return &EjemploService{repo: repo}
}

func (s *EjemploService) CreateEjemplo(ctx context.Context, ejemplo *domain.Ejemplo) error {
    // Validaciones
    // Lógica de negocio
    return s.repo.Create(ctx, ejemplo)
}
```

4. **Implementar el repositorio** en `/internal/adapters/repositories/`
```go
// ejemplo_repository_mongo.go
package repositories

type EjemploRepositoryMongo struct {
    collection *mongo.Collection
}

func NewEjemploRepositoryMongo(collection *mongo.Collection) *EjemploRepositoryMongo {
    return &EjemploRepositoryMongo{collection: collection}
}

func (r *EjemploRepositoryMongo) Create(ctx context.Context, ejemplo *domain.Ejemplo) error {
    result, err := r.collection.InsertOne(ctx, ejemplo)
    if err != nil {
        return err
    }
    ejemplo.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}
```

5. **Crear el handler** en `/internal/adapters/handlers/`
```go
// ejemplo_handler.go
package handlers

type EjemploHandler struct {
    service *services.EjemploService
}

func NewEjemploHandler(service *services.EjemploService) *EjemploHandler {
    return &EjemploHandler{service: service}
}

func (h *EjemploHandler) CreateEjemplo(c *gin.Context) {
    var ejemplo domain.Ejemplo
    if err := c.ShouldBindJSON(&ejemplo); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.service.CreateEjemplo(c.Request.Context(), &ejemplo); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, ejemplo)
}

func (h *EjemploHandler) RegisterRoutes(router *gin.Engine) {
    ejemplos := router.Group("/api/ejemplos")
    {
        ejemplos.POST("", h.CreateEjemplo)
    }
}
```

6. **Registrar en main.go**
```go
// En cmd/api/main.go
ejemploCollection := db.Collection("ejemplos")
ejemploRepo := repositories.NewEjemploRepositoryMongo(ejemploCollection)
ejemploService := services.NewEjemploService(ejemploRepo)
ejemploHandler := handlers.NewEjemploHandler(ejemploService)
ejemploHandler.RegisterRoutes(router)
```

## Comandos Útiles

### Desarrollar
```bash
# Ejecutar aplicación
go run cmd/api/main.go

# Ejecutar con hot reload (instalar: go install github.com/cosmtrek/air@latest)
air

# Ver dependencias
go mod graph

# Descargar dependencias
go mod download
```

### Testing
```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con cobertura
go test -cover ./...

# Ejecutar tests con reporte detallado
go test -v ./...
```

### Formato y Lint
```bash
# Formatear código
go fmt ./...

# Ver problemas de linting (instalar: golangci-lint)
golangci-lint run

# Ejecutar vet (verificador estático built-in)
go vet ./...
```

### Build
```bash
# Compilar para el sistema actual
go build -o bin/api cmd/api/main.go

# Compilar para múltiples plataformas
GOOS=linux GOARCH=amd64 go build -o bin/api-linux cmd/api/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/api-mac cmd/api/main.go
GOOS=windows GOARCH=amd64 go build -o bin/api.exe cmd/api/main.go
```

## Testing con cURL

### Crear Usuario
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dev@example.com",
    "password": "dev123456",
    "role": "client"
  }' | jq
```

### Crear Producto
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Producto de Desarrollo",
    "description": "Producto para testing",
    "base_price": 99.99,
    "category": "Testing",
    "brand": "DevBrand",
    "images": ["https://example.com/image.jpg"],
    "variants": [
      {
        "sku": "DEV-PROD-001",
        "color": "Blue",
        "size": "M",
        "stock": 100,
        "price_adjustment": 0
      }
    ]
  }' | jq
```

## Documentación Local

- 📖 `README.md` - Guía de inicio rápido
- 📐 `ARCHITECTURE.md` - Detalles de arquitectura
- 🔌 `API_ENDPOINTS.md` - Referencia de endpoints
- 📋 `PROJECT_SUMMARY.md` - Resumen del proyecto

## Debugging

### Usar Print Debugging
```go
import "fmt"

fmt.Printf("Valor de x: %v\n", x)
fmt.Printf("Tipo de x: %T\n", x)
```

### Usar Logger
```go
import "log"

log.Printf("Mensaje con valor: %v", variable)
log.Fatal("Error fatal:", err)
```

### Usar Debugger en GoLand
1. Poner breakpoint (click en la línea)
2. Run → Debug 'main'
3. Usar ventana de variables para inspeccionar

## Resolver Errores Comunes

### MongoDB Connection Refused
```
Error: connection refused

Solución:
- Verificar que MongoDB está corriendo
- Verificar MONGO_URI en .env
- Revisar puerto 27017
```

### Import Not Found
```
Error: cannot find module

Solución:
go mod tidy
go mod download
```

### Port Already in Use
```
Error: listen tcp :8080: bind: address already in use

Solución:
# Encontrar proceso en puerto 8080
netstat -ano | findstr :8080

# Matar proceso (Windows)
taskkill /PID <PID> /F

# O cambiar puerto en .env
PORT=8081
```

## Best Practices

✅ Siempre usar contextos en operaciones de base de datos  
✅ Validar entrada en handlers  
✅ Usar interfaces para desacoplamiento  
✅ Documentar funciones públicas  
✅ Manejar errores explícitamente  
✅ Usar logging para debugging  
✅ Escribir tests para lógica de negocio  
✅ Mantener layers separadas y no mezclarlas  

## Resources

- Go Documentation: https://golang.org/doc
- Gin Framework: https://gin-gonic.com
- MongoDB Go Driver: https://pkg.go.dev/go.mongodb.org/mongo-driver
- Clean Architecture: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
