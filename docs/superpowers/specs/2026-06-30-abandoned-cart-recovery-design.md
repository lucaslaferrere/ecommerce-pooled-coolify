# Abandoned Cart Recovery — Design

**Date:** 2026-06-30
**Status:** Approved (pending final spec review)
**Scope:** Fase 2 de la sección Usuarios/Carritos del panel admin.

## Objetivo

Permitir que el admin vea los carritos de clientes logueados que no llegaron a
comprar, y les envíe un email de recuperación con un cupón de descuento elegido
de la lista de cupones existentes.

## Contexto

Hoy el carrito vive solo en el navegador del cliente (Zustand + `persist` →
localStorage, en `poolside-commerce-hub/src/store/cart.ts`). El servidor no lo ve
hasta el checkout. Para recuperar carritos abandonados hay que persistir el
carrito server-side, atado al usuario.

Piezas ya existentes que se reutilizan:
- Sistema de cupones con límite de usos total y por cliente (`max_uses_per_user`).
- `EmailService` (Resend) con patrón `SendOwnerNewOrder` / `SendOwnerNewUser`.
- Sección Usuarios del panel admin (Fase 1).

## Decisiones tomadas

1. **Solo usuarios logueados.** Los invitados arman carrito sin identificarse; no
   hay email para contactarlos. Quedan fuera de scope.
2. **Opción A — sin reloj.** El panel muestra todos los carritos con productos no
   convertidos. No hay lógica de "X horas sin actividad". El admin decide a quién
   enviar.
3. **Sección propia "Carritos"** en el sidebar admin (no dentro de Usuarios).
4. **El carrito server-side se vacía cuando el cliente compra.** Por definición,
   todo lo que queda en la colección `carts` es un carrito no convertido.
5. **Template del mail editable** — solo el cuerpo de texto, no HTML crudo. Con
   variables interpolables, preview en vivo, y default restaurable. Un único
   template global guardado en el backend.
6. **Registrar "última vez enviado"** por carrito para evitar reenvíos accidentales.
7. **Envío manual.** El admin dispara el mail desde el panel. Sin cron/automático.

## Arquitectura

### 1. Persistencia del carrito

**Colección Mongo `carts`** (un documento por usuario):
- `user_id` (ObjectID, único)
- `items` ([]CartItem — mismos ítems que ya usa el checkout)
- `updated_at` (time)
- `last_reminder_sent_at` (time, opcional)

**Endpoints protegidos (cliente logueado):**
- `PUT /cart` — upsert del carrito del usuario autenticado con los ítems actuales.
- `DELETE /cart` — vacía el carrito (se llama en checkout exitoso; también
  disponible si el cliente vacía el carrito manualmente).

**Sincronización en el frontend:**
- En el store del carrito (`cart.ts`), cuando cambian los `items` y hay usuario
  logueado, hacer `PUT /cart` con debounce (~2s tras el último cambio).
- Al iniciar sesión, **sube el carrito local** (PUT) — el carrito local es la
  fuente de verdad porque es donde el cliente está trabajando. No se hace merge
  con lo que hubiera en el server (mantiene el flujo simple y predecible).
- En checkout exitoso, el carrito local ya se limpia (`clear()`); agregar el
  borrado server-side (`DELETE /cart` o vaciado dentro del flujo de checkout).

### 2. Vista en el panel

**Sección "Carritos"** (`/admin/carts`, link en `AdminLayout`).

**Endpoint admin:** `GET /admin/carts` — lista de carritos no vacíos, enriquecidos
con el email del usuario. Devuelve: user_id, email, cantidad de ítems, total,
updated_at, last_reminder_sent_at.

**UI:**
- Tabla: email, # de ítems, total, última actualización, (indicador si ya se
  envió recordatorio).
- Botón "Ver" → modal con los productos y cantidades del carrito.
- Botón "Enviar cupón" → flujo de envío (sección 3).

### 3. Envío del mail con cupón

**Flujo:**
1. El admin elige un cupón de un dropdown (cupones activos).
2. Ve un preview del mail con las variables ya interpoladas (código, descuento,
   vencimiento) + la tabla de productos del carrito.
3. Confirma → el backend arma el HTML y lo envía al email del cliente.
4. Se registra `last_reminder_sent_at` en el carrito.

**Endpoint admin:** `POST /admin/carts/:userID/send-coupon` con body `{ coupon_id }`.

**Template editable:**
- Se guarda un único template global en el backend: un documento en la colección
  `settings` con key `abandoned_cart_email` y un campo `body` (texto).
- El admin edita solo el **cuerpo de texto**. El backend lo envuelve en la
  plantilla HTML (header, botón CTA, tabla de productos, footer).
- **Variables disponibles:** `{{codigo}}`, `{{descuento}}`, `{{vencimiento}}`,
  `{{productos}}`, `{{total}}`. (Nombre del cliente no está disponible: el modelo
  User solo tiene email — se puede usar el email o un saludo genérico.)
- Endpoints admin para el template: `GET /admin/settings/abandoned-cart-email`,
  `PUT /admin/settings/abandoned-cart-email`. Botón "restaurar default" en la UI
  vuelve al texto por defecto embebido en el código.
- Editor accesible desde la sección Carritos ("Editar plantilla del mail").

## Fuera de scope (YAGNI)

- Captura de carritos de invitados.
- Envío automático / cron / filtro por tiempo de inactividad.
- Edición de HTML crudo del mail.
- Múltiples templates o templates por cupón.

## Riesgos / notas

- El "total" del carrito en el panel es orientativo: los precios reales se
  recalculan server-side en el checkout. Mostrar el total con los `unit_price`
  guardados en el carrito.
- Un carrito puede quedar desactualizado si el cliente limpia su localStorage sin
  loguearse de nuevo; se sincroniza en el próximo login/cambio.
- El cupón sugerido para recuperación debería tener `max_uses_per_user = 1` para
  blindarlo, pero eso lo controla el admin al crear el cupón — no se fuerza aquí.
