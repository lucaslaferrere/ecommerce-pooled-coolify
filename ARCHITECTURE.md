# Ecommerce Backend - Clean Architecture

## Estructura del Proyecto

```
ecommerce-pooled/
├── cmd/
│   └── api/                    # Punto de entrada de la aplicación
├── internal/
│   ├── adapters/              # Adaptadores (implementaciones concretas)
│   │   ├── database/          # Configuración de base de datos
│   │   ├── handlers/          # HTTP handlers (Gin)
│   │   └── repositories/      # Implementaciones de repositorios
│   ├── config/                # Configuración de la aplicación
│   ├── core/                  # Lógica de negocio (Clean Architecture)
│   │   ├── domain/            # Entidades de dominio (puros)
│   │   ├── ports/             # Interfaces (contratos)
│   │   └── services/          # Lógica de negocio
│   └── utils/                 # Utilidades generales
├── go.mod
├── go.sum
└── README.md
```

## Estructura de Clean Architecture / Hexagonal

### Core (Núcleo de Negocio)
- **Domain**: Entidades puras sin dependencias externas
- **Ports**: Interfaces que definen contratos para adaptadores
- **Services**: Casos de uso y lógica de negocio

### Adapters (Implementaciones Concretas)
- **Handlers**: HTTP controllers usando Gin
- **Repositories**: Implementaciones de persistencia con MongoDB
- **Database**: Configuración de conexión a MongoDB

## Entidades de Dominio

### User
- ID (ObjectID de MongoDB)
- Email
- PasswordHash
- Role (admin | client)

### Product
- ID (ObjectID de MongoDB)
- Name
- Description
- BasePrice
- Category
- Brand
- Images (array de strings)
- Variants (slice de Variant)
  - SKU
  - Color
  - Size
  - Stock
  - PriceAdjustment

### Order
- ID (ObjectID de MongoDB)
- UserID
- Items (slice de CartItem)
- Total
- Status (pending | processing | shipped | delivered | cancelled)
- CreatedAt, UpdatedAt

### CartItem
- ProductID
- VariantSKU
- Quantity

## Flujo de Dependencias

```
HTTP Request → Handlers (Adapters) → Services (Core) → Repositories (Ports)
                                                              ↓
                                                         Database (MongoDB)
```

## Tecnologías

- **Lenguaje**: Go 1.25.6
- **Framework HTTP**: Gin
- **Base de Datos**: MongoDB (mongo-go-driver)
- **Arquitectura**: Clean Architecture / Hexagonal
