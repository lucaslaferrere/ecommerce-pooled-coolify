// Package tests contains end-to-end integration tests that spin up the full
// Gin router (same wiring as production) against a real MongoDB instance.
//
// Requirements:
//   - MongoDB must be reachable. The tests read MONGO_URI (default: mongodb://localhost:27017).
//   - If MongoDB is unavailable the tests skip automatically — they never fail due to connectivity.
//
// Run:
//
//	go test ./tests/... -v -count=1
//	MONGO_URI=mongodb://user:pass@host:27017 go test ./tests/... -v -count=1
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ecommerce-pooled/internal/adapters/database"
	"ecommerce-pooled/internal/app"
	"ecommerce-pooled/internal/config"
	"ecommerce-pooled/internal/core/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ── Test infrastructure ───────────────────────────────────────────────────────

// suite holds the router and a handle to the test database so individual tests
// can query MongoDB directly to assert side-effects (e.g. stock decremented).
type suite struct {
	router *gin.Engine
	db     *mongo.Database
}

// newSuite connects to MongoDB, builds the router and registers cleanup.
// It skips (never fails) when MongoDB is not reachable.
func newSuite(t *testing.T) *suite {
	t.Helper()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Use an isolated test database; a timestamp suffix avoids cross-run collisions
	// when running tests in parallel or without -count=1.
	dbName := fmt.Sprintf("ecommerce_test_%d", time.Now().UnixNano())

	client, err := database.NewClient(context.Background(), mongoURI, dbName)
	if err != nil {
		t.Skipf("integration tests require MongoDB (%s): %v", mongoURI, err)
	}

	db := client.GetDatabase()

	cfg := &config.Config{
		JWTSecret:       "integration-test-secret",
		AppURL:          "http://localhost:8080",
		GinMode:         gin.TestMode,
		ShutdownTimeout: 5,
	}
	gin.SetMode(gin.TestMode)

	router := app.BuildRouter(cfg, db)

	t.Cleanup(func() {
		// Drop the entire test database so nothing leaks between runs
		_ = db.Drop(context.Background())
		_ = client.Close(context.Background())
	})

	return &suite{router: router, db: db}
}

// do sends an HTTP request through the router and returns the recorder.
func (s *suite) do(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// mustParseBody decodes the response body into v; fails the test on error.
func mustParseBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	require.NoError(t, json.NewDecoder(w.Body).Decode(v))
}

// doMultipart sends a multipart/form-data request. fields are plain text values.
func (s *suite) doMultipart(t *testing.T, method, path string, fields map[string]string, token string) *httptest.ResponseRecorder {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// ── Auth helpers ──────────────────────────────────────────────────────────────

type credentials struct {
	Email    string
	Password string
	Role     string
}

// registerAndLogin creates a user via the API and returns the access token.
func (s *suite) registerAndLogin(t *testing.T, creds credentials) string {
	t.Helper()

	// Register
	w := s.do(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    creds.Email,
		"password": creds.Password,
		"role":     creds.Role,
	}, "")
	require.Equal(t, http.StatusCreated, w.Code,
		"register failed: %s", w.Body.String())

	// Login
	w = s.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    creds.Email,
		"password": creds.Password,
	}, "")
	require.Equal(t, http.StatusOK, w.Code,
		"login failed: %s", w.Body.String())

	var resp map[string]interface{}
	mustParseBody(t, w, &resp)

	token, ok := resp["access_token"].(string)
	require.True(t, ok, "access_token missing from login response")
	return token
}

// ── Test 1: Admin product CRUD ────────────────────────────────────────────────

// TestAdminProductCRUD exercises the full admin lifecycle for a product:
// register → login → create → update price → list (verify present) → delete → list (verify gone).
func TestAdminProductCRUD(t *testing.T) {
	s := newSuite(t)

	// 1. Register admin and obtain JWT
	adminToken := s.registerAndLogin(t, credentials{
		Email:    "admin@test.com",
		Password: "secret123",
		Role:     "admin",
	})

	// 2. Create a product (handler expects multipart/form-data)
	t.Run("create product", func(t *testing.T) {
		variantsJSON := `[{"sku":"RX1-BLK-42","color":"Negro","size":"42","stock":10,"price_adjustment":0},{"sku":"RX1-WHT-42","color":"Blanco","size":"42","stock":5,"price_adjustment":10}]`
		fields := map[string]string{
			"name":        "Zapatilla Runner X1",
			"description": "Zapatilla de running de alta performance",
			"base_price":  "150.00",
			"category":    "calzado",
			"brand":       "RunPro",
			"variants":    variantsJSON,
		}

		w := s.doMultipart(t, http.MethodPost, "/api/v1/admin/products", fields, adminToken)
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		var created map[string]interface{}
		mustParseBody(t, w, &created)
		assert.NotEmpty(t, created["id"])
		assert.Equal(t, "Zapatilla Runner X1", created["name"])

		productID := created["id"].(string)

		// 3. Update the product price (also multipart/form-data)
		t.Run("update price", func(t *testing.T) {
			w := s.doMultipart(t, http.MethodPut, "/api/v1/admin/products/"+productID,
				map[string]string{"base_price": "175.00"}, adminToken)
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

			var updated map[string]interface{}
			mustParseBody(t, w, &updated)
			assert.Equal(t, 175.00, updated["base_price"])
		})

		// 4. List products and verify the product is present with the new price
		t.Run("list products", func(t *testing.T) {
			w := s.do(t, http.MethodGet, "/api/v1/products", nil, "")
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

			var resp map[string]interface{}
			mustParseBody(t, w, &resp)

			products := resp["products"].([]interface{})
			require.Len(t, products, 1, "expected exactly 1 product")

			p := products[0].(map[string]interface{})
			assert.Equal(t, "Zapatilla Runner X1", p["name"])
			assert.Equal(t, 175.00, p["base_price"])
		})

		// 5. Delete the product
		t.Run("delete product", func(t *testing.T) {
			w := s.do(t, http.MethodDelete, "/api/v1/admin/products/"+productID, nil, adminToken)
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		})

		// 6. List again — catalog should be empty
		t.Run("product gone after delete", func(t *testing.T) {
			w := s.do(t, http.MethodGet, "/api/v1/products", nil, "")
			assert.Equal(t, http.StatusOK, w.Code)

			var resp map[string]interface{}
			mustParseBody(t, w, &resp)

			products, _ := resp["products"].([]interface{}) // nil when empty (JSON null)
			assert.Len(t, products, 0, "catalog should be empty after deletion")
		})
	})
}

// ── Test 2: Client checkout flow ──────────────────────────────────────────────

// TestClientCheckoutFlow exercises the shopper journey:
// admin creates product → client registers → client checks out →
// verify order status "pending" + shipping details stored →
// verify stock decremented in DB →
// client lists their orders.
func TestClientCheckoutFlow(t *testing.T) {
	s := newSuite(t)

	// ── Setup: admin creates a product the client will buy ────────────────────
	adminToken := s.registerAndLogin(t, credentials{
		Email:    "admin2@test.com",
		Password: "admin123",
		Role:     "admin",
	})

	w := s.doMultipart(t, http.MethodPost, "/api/v1/admin/products", map[string]string{
		"name":        "Remera Básica",
		"description": "100% algodón",
		"base_price":  "25.00",
		"category":    "ropa",
		"brand":       "BasicCo",
		"variants":    `[{"sku":"REM-WHT-M","color":"Blanco","size":"M","stock":10,"price_adjustment":0}]`,
	}, adminToken)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var createdProduct map[string]interface{}
	mustParseBody(t, w, &createdProduct)
	productID := createdProduct["id"].(string)

	// ── Client registers and logs in ──────────────────────────────────────────
	clientToken := s.registerAndLogin(t, credentials{
		Email:    "client@test.com",
		Password: "pass1234",
		Role:     "client",
	})

	// ── Checkout ──────────────────────────────────────────────────────────────
	t.Run("checkout creates pending order", func(t *testing.T) {
		checkoutPayload := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id":  productID,
					"variant_sku": "REM-WHT-M",
					"quantity":    3,
				},
			},
			"customer_name":    "Juan Pérez",
			"customer_email":   "juan@test.com",
			"customer_phone":   "1112345678",
			"shipping_address": "Av. Corrientes 1234",
			"shipping_city":    "Buenos Aires",
			"shipping_zip":     "C1043",
		}

		w := s.do(t, http.MethodPost, "/api/v1/checkout", checkoutPayload, clientToken)
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		// Checkout retorna {order, init_point, sandbox_init_point}
		var resp map[string]interface{}
		mustParseBody(t, w, &resp)

		// init_point puede ser vacío cuando MP_ACCESS_TOKEN no está configurado
		_, hasInitPoint := resp["init_point"]
		assert.True(t, hasInitPoint, "respuesta debe incluir la clave init_point")

		order := resp["order"].(map[string]interface{})

		// Order fields
		assert.Equal(t, "pending", order["status"])
		assert.Equal(t, 75.00, order["total"]) // 25.00 * 3

		// Shipping details persisted
		sd := order["shipping_details"].(map[string]interface{})
		assert.Equal(t, "Av. Corrientes 1234", sd["address"])
		assert.Equal(t, "Buenos Aires", sd["city"])
		assert.Equal(t, "C1043", sd["postal_code"])

		orderID := order["id"].(string)

		// ── Verify stock was decremented in MongoDB ───────────────────────────
		t.Run("variant stock decremented", func(t *testing.T) {
			oid, err := primitive.ObjectIDFromHex(productID)
			require.NoError(t, err)

			var product domain.Product
			err = s.db.Collection("products").
				FindOne(context.Background(), bson.M{"_id": oid}).
				Decode(&product)
			require.NoError(t, err)

			var found bool
			for _, v := range product.Variants {
				if v.SKU == "REM-WHT-M" {
					assert.Equal(t, 7, v.Stock, // started at 10, bought 3 → 7
						"expected stock=7 after buying qty=3 from stock=10")
					found = true
					break
				}
			}
			assert.True(t, found, "variant REM-WHT-M not found in product")
		})

		// ── Order appears in GET /orders/me ───────────────────────────────────
		t.Run("order visible in my orders", func(t *testing.T) {
			w := s.do(t, http.MethodGet, "/api/v1/orders/me", nil, clientToken)
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

			var resp map[string]interface{}
			mustParseBody(t, w, &resp)

			orders := resp["orders"].([]interface{})
			require.GreaterOrEqual(t, len(orders), 1, "expected at least 1 order")

			// Find our order by ID
			var found bool
			for _, o := range orders {
				om := o.(map[string]interface{})
				if om["id"] == orderID {
					assert.Equal(t, "pending", om["status"])
					found = true
					break
				}
			}
			assert.True(t, found, "checkout order %s not found in /orders/me", orderID)
		})
	})

	// ── Checkout rejects insufficient stock ───────────────────────────────────
	t.Run("checkout fails when stock is insufficient", func(t *testing.T) {
		checkoutPayload := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id":  productID,
					"variant_sku": "REM-WHT-M",
					"quantity":    999,
				},
			},
			"customer_name":    "Ana García",
			"customer_email":   "ana@test.com",
			"customer_phone":   "1198765432",
			"shipping_address": "Calle Falsa 123",
			"shipping_city":    "Córdoba",
			"shipping_zip":     "5000",
		}

		w := s.do(t, http.MethodPost, "/api/v1/checkout", checkoutPayload, clientToken)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var resp map[string]interface{}
		mustParseBody(t, w, &resp)
		assert.NotEmpty(t, resp["error"])
	})

	// ── Admin can change order status ─────────────────────────────────────────
	t.Run("admin updates order status to processing", func(t *testing.T) {
		// Re-checkout a small order so we have a fresh order ID
		payload := map[string]interface{}{
			"items": []map[string]interface{}{
				{"product_id": productID, "variant_sku": "REM-WHT-M", "quantity": 1},
			},
			"customer_name":    "Test User",
			"customer_email":   "test@test.com",
			"customer_phone":   "1100000000",
			"shipping_address": "Test 1",
			"shipping_city":    "CABA",
			"shipping_zip":     "1000",
		}
		w := s.do(t, http.MethodPost, "/api/v1/checkout", payload, clientToken)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

		var checkoutResp map[string]interface{}
		mustParseBody(t, w, &checkoutResp)
		orderID := checkoutResp["order"].(map[string]interface{})["id"].(string)

		w = s.do(t, http.MethodPatch, "/api/v1/admin/orders/"+orderID+"/status",
			map[string]string{"status": "processing"}, adminToken)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})
}
