#!/bin/bash
# 🔍 VERIFICACIÓN DEL PROYECTO - Ecommerce Backend Go
# Este script verifica que todos los archivos y estructura estén correctos

echo "=========================================="
echo "VERIFICACIÓN DEL PROYECTO"
echo "Ecommerce Backend - Clean Architecture"
echo "=========================================="
echo ""

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Contador
CHECKS=0
PASSED=0

# Función para verificar archivo
check_file() {
    CHECKS=$((CHECKS + 1))
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗${NC} $1 (FALTA)"
    fi
}

# Función para verificar directorio
check_dir() {
    CHECKS=$((CHECKS + 1))
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗${NC} $1/ (FALTA)"
    fi
}

echo "📁 Verificando Directorios..."
echo "----"
check_dir "cmd/api"
check_dir "internal/adapters"
check_dir "internal/adapters/database"
check_dir "internal/adapters/handlers"
check_dir "internal/adapters/repositories"
check_dir "internal/config"
check_dir "internal/core"
check_dir "internal/core/domain"
check_dir "internal/core/ports"
check_dir "internal/core/services"
check_dir "internal/utils"
echo ""

echo "📄 Verificando Archivos Go (Domain)..."
echo "----"
check_file "internal/core/domain/user.go"
check_file "internal/core/domain/product.go"
check_file "internal/core/domain/cart_item.go"
check_file "internal/core/domain/order.go"
echo ""

echo "📄 Verificando Archivos Go (Ports)..."
echo "----"
check_file "internal/core/ports/user_repository.go"
check_file "internal/core/ports/product_repository.go"
check_file "internal/core/ports/order_repository.go"
echo ""

echo "📄 Verificando Archivos Go (Services)..."
echo "----"
check_file "internal/core/services/user_service.go"
check_file "internal/core/services/product_service.go"
check_file "internal/core/services/order_service.go"
check_file "internal/core/services/errors.go"
echo ""

echo "📄 Verificando Archivos Go (Handlers)..."
echo "----"
check_file "internal/adapters/handlers/user_handler.go"
check_file "internal/adapters/handlers/product_handler.go"
check_file "internal/adapters/handlers/order_handler.go"
echo ""

echo "📄 Verificando Archivos Go (Repositories)..."
echo "----"
check_file "internal/adapters/repositories/user_repository_mongo.go"
check_file "internal/adapters/repositories/product_repository_mongo.go"
check_file "internal/adapters/repositories/order_repository_mongo.go"
echo ""

echo "📄 Verificando Archivos Go (Otros)..."
echo "----"
check_file "internal/adapters/database/mongo.go"
check_file "internal/config/config.go"
check_file "internal/utils/password.go"
check_file "cmd/api/main.go"
echo ""

echo "📄 Verificando Documentación..."
echo "----"
check_file "go.mod"
check_file "go.sum"
check_file ".env.example"
check_file "README.md"
check_file "ARCHITECTURE.md"
check_file "DEVELOPMENT.md"
check_file "API_ENDPOINTS.md"
check_file "PROJECT_SUMMARY.md"
check_file "COMPLETION_REPORT.md"
check_file "FILES_STRUCTURE.md"
check_file "INDEX.md"
echo ""

# Verificar compilación
echo "🔨 Verificando Compilación..."
echo "----"
CHECKS=$((CHECKS + 1))
if go build -o /tmp/test-build cmd/api/main.go 2>/dev/null; then
    echo -e "${GREEN}✓${NC} go build exitoso"
    PASSED=$((PASSED + 1))
    rm /tmp/test-build
else
    echo -e "${RED}✗${NC} go build falló"
fi
echo ""

# Resumen
echo "=========================================="
echo "RESUMEN"
echo "=========================================="
PERCENTAGE=$((PASSED * 100 / CHECKS))
echo "Verificaciones: $PASSED/$CHECKS ($PERCENTAGE%)"
echo ""

if [ $PERCENTAGE -eq 100 ]; then
    echo -e "${GREEN}✓ PROYECTO VERIFICADO EXITOSAMENTE${NC}"
    echo "El proyecto está 100% completo."
    exit 0
else
    echo -e "${YELLOW}⚠ PROYECTO INCOMPLETO${NC}"
    echo "Faltan $((CHECKS - PASSED)) archivos/directorios."
    exit 1
fi
