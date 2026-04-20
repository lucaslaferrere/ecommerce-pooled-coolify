# 🚀 Quick Reference - Comandos Esenciales

## 📋 Tabla de Contenidos Rápida

| Categoría | Comando | Descripción |
|-----------|---------|------------|
| **Setup** | `go mod download` | Descargar dependencias |
| **Setup** | `go mod tidy` | Limpiar dependencias |
| **Desarrollo** | `go run cmd/api/main.go` | Ejecutar aplicación |
| **Build** | `go build -o bin/api cmd/api/main.go` | Compilar binario |
| **Testing** | `go test ./...` | Ejecutar tests |
| **Formato** | `go fmt ./...` | Formatear código |
| **Lint** | `go vet ./...` | Verificar código |
| **MongoDB** | `docker run -d -p 27017:27017 mongo:latest` | Iniciar MongoDB en Docker |

---

## 🔧 Setup Inicial

```bash
# 1. Navegar al proyecto
cd E:\LyRSolutions\ecommerce-pooled

# 2. Descargar dependencias
go mod download
go mod tidy

# 3. Copiar configuración de ejemplo
cp .env.example .env

# 4. Iniciar MongoDB (en otra terminal)
docker run -d -p 27017:27017 mongo:latest

# 5. Ejecutar aplicación
go run cmd/api/main.go

# 6. Probar en otra terminal
curl http://localhost:8080/health
```

---

## 🏃 Comandos de Desarrollo

### Ejecutar Aplicación
```bash
# Desarrollo normal
go run cmd/api/main.go

# Con hot reload (requiere instalar air)
air
```

### Compilar
```bash
# Windows
go build -o bin/api.exe cmd/api/main.go

# Linux/Mac
go build -o bin/api cmd/api/main.go

# Cross-compile para Linux
GOOS=linux GOARCH=amd64 go build -o bin/api-linux cmd/api/main.go
```

### Ejecutar Binario Compilado
```bash
# Windows
.\bin\api.exe

# Linux/Mac
./bin/api
```

---

## 🧪 Testing

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con output verbose
go test -v ./...

# Ver cobertura de tests
go test -cover ./...

# Generar reporte de cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🔍 Análisis de Código

```bash
# Formatear código
go fmt ./...

# Ver problemas potenciales (go vet)
go vet ./...

# Ejecutar linter (requiere golangci-lint)
golangci-lint run

# Ver dependencias del módulo
go mod graph

# Ver qué packages dependen de algo
go mod why -m <package>
```

---

## 📦 Gestión de Dependencias

```bash
# Descargar todas las dependencias
go mod download

# Limpiar dependencias no usadas
go mod tidy

# Verificar integridad
go mod verify

# Actualizar dependencia específica
go get -u github.com/gin-gonic/gin@latest

# Actualizar todas las dependencias
go get -u ./...
```

---

## 🧬 MongoDB

### Iniciar MongoDB en Docker
```bash
# Crear y ejecutar contenedor
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:latest

# Ver logs
docker logs mongodb

# Detener MongoDB
docker stop mongodb

# Iniciar MongoDB detenido
docker start mongodb

# Eliminar MongoDB
docker rm mongodb
```

### Conectar a MongoDB desde Shell
```bash
# Usando mongosh (nuevo)
docker exec -it mongodb mongosh -u admin -p password

# Usar mongo shell (deprecated)
docker exec -it mongodb mongo -u admin -p password
```

### Comandos MongoDB Básicos
```bash
# Listar bases de datos
show dbs

# Usar base de datos
use ecommerce

# Listar colecciones
show collections

# Ver documentos de users
db.users.find().pretty()

# Contar documentos
db.users.countDocuments()

# Insertar documento
db.users.insertOne({
  email: "test@example.com",
  role: "client"
})
```

---

## 📡 Testing API con cURL

### Crear Usuario
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "role": "client"
  }'
```

### Obtener Usuario
```bash
curl http://localhost:8080/api/users/507f1f77bcf86cd799439011
```

### Crear Producto
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "Professional Laptop",
    "base_price": 999.99,
    "category": "Electronics",
    "brand": "Dell",
    "images": ["url1", "url2"],
    "variants": [
      {
        "sku": "LAPTOP-001",
        "color": "Silver",
        "size": "15",
        "stock": 50,
        "price_adjustment": 0
      }
    ]
  }'
```

### Crear Orden
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "507f1f77bcf86cd799439011",
    "items": [
      {
        "product_id": "507f1f77bcf86cd799439012",
        "variant_sku": "LAPTOP-001",
        "quantity": 1
      }
    ],
    "total": 999.99,
    "status": "pending"
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

---

## 🐛 Debugging

### Usar delve (debugger de Go)
```bash
# Instalar delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Ejecutar con debugger
dlv debug cmd/api/main.go

# Comandos en delve
help              # Ver ayuda
break main.main   # Poner breakpoint
continue          # Continuar
next              # Siguiente línea
step              # Entrar en función
print var_name    # Ver variable
```

### Print Debugging
```go
import "fmt"

// En el código
fmt.Printf("Valor: %v\n", variable)
fmt.Printf("Tipo: %T\n", variable)
```

### Usar Logging
```go
import "log"

log.Printf("Debug: %v", value)
log.Fatalf("Error fatal: %v", err)
```

---

## 🛠️ Troubleshooting Rápido

### Error: "connection refused"
```bash
# Verificar que MongoDB está corriendo
docker ps | grep mongodb

# O iniciar MongoDB
docker run -d -p 27017:27017 mongo:latest
```

### Error: "port already in use"
```bash
# Ver qué está usando el puerto (Windows)
netstat -ano | findstr :8080

# Matar proceso (reemplazar PID)
taskkill /PID <PID> /F

# O cambiar puerto en .env
# PORT=8081
```

### Error: "import not found"
```bash
# Descargar dependencias
go mod download

# Limpiar go.mod
go mod tidy
```

### Binario no ejecutable
```bash
# Hacer ejecutable (Linux/Mac)
chmod +x bin/api

# Ejecutar
./bin/api
```

---

## 📊 Performance

### Ver estadísticas de compilación
```bash
go build -v cmd/api/main.go
```

### Benchmark (si hay tests)
```bash
go test -bench=. ./...
go test -benchmem ./...
```

### Profile de memoria
```bash
# Crear profile
go build -o bin/api cmd/api/main.go
# En otro terminal con aplicación corriendo
# go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 📚 Documentación Rápida

### Dentro del Proyecto
- `README.md` - Inicio rápido
- `DEVELOPMENT.md` - Desarrollo detallado
- `API_ENDPOINTS.md` - Referencia de API
- `ARCHITECTURE.md` - Arquitectura
- `INDEX.md` - Índice completo

### Online
- Go: https://golang.org/doc
- Gin: https://gin-gonic.com
- MongoDB Go: https://pkg.go.dev/go.mongodb.org/mongo-driver

---

## 🎯 Flujo Típico de Desarrollo

```bash
# 1. Hacer cambios al código
# (editar archivos)

# 2. Formatear código
go fmt ./...

# 3. Verificar código
go vet ./...

# 4. Ejecutar aplicación
go run cmd/api/main.go

# 5. Probar con cURL
curl http://localhost:8080/health

# 6. Si todo OK, compilar
go build -o bin/api cmd/api/main.go

# 7. Ejecutar binario
./bin/api
```

---

## 🔐 Seguridad

### Verificar vulnerabilidades
```bash
go list -json -m all | nancy sleuth
```

### Actualizar dependencias
```bash
go get -u ./...
go mod tidy
```

---

## 📈 Métricas Útiles

### Tamaño del binario
```bash
ls -lh bin/api
```

### Número de líneas de código
```bash
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l
```

### Dependencies
```bash
go mod graph | wc -l
```

---

## 💾 Comandos Git (si usas control de versión)

```bash
# Inicializar repositorio
git init

# Agregar archivos
git add .

# Commit
git commit -m "Tarea 1: Setup proyecto Clean Architecture"

# Ver estado
git status

# Ver historial
git log
```

---

## ✨ Atajo: Script de Inicio Rápido

Crea un archivo `start.sh`:

```bash
#!/bin/bash

echo "🚀 Iniciando Ecommerce Backend..."

# 1. Descargar dependencias
echo "📦 Descargando dependencias..."
go mod download

# 2. MongoDB
echo "🗄️ Iniciando MongoDB..."
docker run -d -p 27017:27017 mongo:latest

# 3. Esperar a que MongoDB esté listo
sleep 5

# 4. Ejecutar aplicación
echo "▶️ Ejecutando aplicación..."
go run cmd/api/main.go

echo "✅ Listo en http://localhost:8080"
```

Uso:
```bash
chmod +x start.sh
./start.sh
```

---

## 🎯 Resumen de Rutas Comunes

| Tarea | Comando |
|-------|---------|
| Empezar | `go run cmd/api/main.go` |
| Build | `go build -o bin/api cmd/api/main.go` |
| Tests | `go test ./...` |
| Formato | `go fmt ./...` |
| Lint | `go vet ./...` |
| MongoDB | `docker run -d -p 27017:27017 mongo:latest` |
| Clean | `go clean` |
| Update deps | `go get -u ./...` |

---

*Quick Reference v1.0 - Ecommerce Backend Go*
