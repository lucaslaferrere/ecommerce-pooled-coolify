# Bulk Price Update — Design

**Date:** 2026-07-02
**Status:** Approved (pending final spec review)
**Scope:** Actualización masiva de precios de productos desde el panel admin, por porcentaje, con selección flexible (todos / por categoría / manual).

## Objetivo

Permitir al admin actualizar el precio de varios productos a la vez aplicando un
porcentaje (aumento o descuento), en lugar de editar producto por producto.

## Decisiones tomadas

1. El porcentaje se aplica al **`base_price`** del producto. Los `price_adjustment`
   de las variantes (extras fijos por color/talle) NO se tocan.
2. El porcentaje puede ser **positivo (aumento) o negativo (descuento)**.
3. El resultado se **redondea a peso entero** (`math.Round`).
4. **Paso de confirmación con preview** en el frontend antes de aplicar: se muestra
   cuántos productos se van a afectar y ejemplos (precio viejo → nuevo).
5. La operación se hace en el backend en **una sola llamada** (endpoint bulk), no
   iterando N requests desde el navegador.

## Contexto (modelo existente)

- `domain.Product.BasePrice float64` — el precio base del producto.
- `domain.Variant.PriceAdjustment float64` — extra fijo por variante (fuera de scope).
- `Category` (frontend) = `'luminarias' | 'controladores' | 'kits' | 'osire' | 'accesorios'`.
  El "tipo" de la selección por categoría usa `Product.category`.
- Página admin de productos: `src/pages/admin/Products.tsx`, `ProductTable.tsx`,
  `ProductFormModal.tsx`.
- Endpoint de update individual ya existe: `PUT /admin/products/:id` (multipart).

## Arquitectura

### Backend

**Endpoint nuevo:** `PATCH /admin/products/bulk-price` (grupo admin, detrás de
auth + admin middleware).

Body:
```json
{ "product_ids": ["<hex>", "..."], "percent": 15 }
```

Semántica:
- Para cada `product_id`: `new = round(base_price * (1 + percent/100))`.
- `percent` puede ser negativo (descuento). Validar rango razonable (ej. entre -90
  y 1000) para evitar destrozos por typo; rechazar si `product_ids` está vacío.
- El precio nunca queda negativo (piso en 0, aunque con rango válido no debería pasar).
- Actualiza `base_price` y `updated_at`. No toca variantes.

**Capas:**
- `ProductService.BulkUpdatePrice(ctx, ids []ObjectID, percent float64) (updated int, err error)`
  — carga cada producto, recalcula, guarda. Devuelve cuántos actualizó.
- `ProductRepository` — reusa `GetByID` + `Update` existentes, o un update de campo
  puntual si ya existe un patrón (`SetBasePrice`-style). Preferir reusar lo que haya.
- `ProductHandler.BulkUpdatePrice` — parsea el body, valida, llama al service,
  responde `{ "updated": N }`.

### Frontend

En la página de productos del admin (`Products.tsx` + `ProductTable.tsx`):

- **Selección:** checkbox por fila; estado de selección (`Set<productId>`) en
  `Products.tsx`.
- **"Seleccionar todos":** tilda todos los productos de la lista actual.
- **"Seleccionar por tipo":** dropdown de categorías; al elegir una, agrega a la
  selección todos los productos de esa categoría.
- **Selección manual:** tildar/destildar filas.
- **Barra de acciones masivas:** aparece cuando hay ≥1 seleccionado; muestra
  "N seleccionados", un input de porcentaje (acepta negativos) y botón "Aplicar".
- **Preview de confirmación:** al apretar "Aplicar", un diálogo muestra la cantidad
  de productos afectados y 2-3 ejemplos (`nombre: $viejo → $nuevo`, calculado en el
  cliente con el mismo redondeo). Confirmar dispara el `PATCH`.
- Al confirmar: `apiPatch('/admin/products/bulk-price', { product_ids, percent })`,
  invalida la query de productos del admin, toast de éxito, limpia la selección.

## Flujo de datos

1. Admin tilda productos (todos / por categoría / manual).
2. Escribe el porcentaje (+/-) y aprieta "Aplicar".
3. Preview: "Vas a actualizar N productos" + ejemplos → confirmar.
4. `PATCH /admin/products/bulk-price` → backend recalcula `base_price` de cada uno.
5. Respuesta `{ updated: N }` → invalida cache, refresca la tabla, toast.

## Manejo de errores

- `product_ids` vacío → 400.
- `percent` fuera de rango (ej. < -90 o > 1000) → 400 con mensaje claro.
- ID inválido en la lista → se ignora ese ID y se cuenta solo lo actualizado (o
  se rechaza toda la operación; elegir: **ignorar inválidos y reportar `updated`**
  para no frenar todo por un ID viejo).
- Falla de DB en un producto → loguear y continuar; `updated` refleja los exitosos.

## Fuera de scope (YAGNI)

- Deshacer / historial de cambios de precio.
- Escalar los `price_adjustment` de variantes.
- Aplicar montos fijos (solo porcentaje).
- Programar cambios de precio a futuro.

## Riesgos / notas

- Cambio masivo e irreversible sin historial: el preview de confirmación es la
  salvaguarda principal. Rango de `percent` acotado como segunda barrera.
- El precio mostrado en la tienda es `base_price (+ price_adjustment por variante)`;
  como no tocamos los ajustes, las variantes con extra mantienen su diferencia.
