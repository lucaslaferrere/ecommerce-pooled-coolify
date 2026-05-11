// Package app wires the full dependency graph and returns a configured Gin router.
// Both cmd/api/main.go and the integration tests use BuildRouter so that tests
// exercise the exact same middleware stack and route table as production.
package app

import (
	"log"
	"net/http"

	"ecommerce-pooled/internal/adapters/handlers"
	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/config"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// BuildRouter wires repositories → services → handlers and registers all routes
// under /api/v1.  gin.Logger() is only added when GinMode != "test" so that
// integration test output stays clean.
func BuildRouter(cfg *config.Config, db *mongo.Database) *gin.Engine {
	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repositories.NewUserRepositoryMongo(db.Collection("users"))
	productRepo := repositories.NewProductRepositoryMongo(db.Collection("products"))
	kitRepo             := repositories.NewKitRepositoryMongo(db.Collection("kits"))
	orderRepo           := repositories.NewOrderRepositoryMongo(db.Collection("orders"))
	distributorLeadRepo := repositories.NewDistributorLeadRepositoryMongo(db.Collection("distributor_leads"))
	wizardRepo := repositories.NewWizardRecommendationRepositoryMongo(db.Collection("wizard_recommendations"))

	// ── Services ─────────────────────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	kitService             := services.NewKitService(kitRepo)
	cartService            := services.NewCartService(productRepo)
	orderService           := services.NewOrderService(orderRepo, productRepo)
	distributorLeadService := services.NewDistributorLeadService(distributorLeadRepo)
	wizardService := services.NewWizardRecommendationService(wizardRepo)
	imageStorage := services.NewLocalImageStorage("./uploads", cfg.AppURL+"/uploads")

	paymentService, err := services.NewPaymentService(
		cfg.MPAccessToken,
		cfg.MPWebhookSecret,
		cfg.MPSuccessURL,
		cfg.MPFailureURL,
		cfg.MPPendingURL,
		cfg.AppURL,
		cfg.MPCurrencyID,
	)
	if err != nil {
		log.Fatalf("error inicializando PaymentService: %v", err)
	}

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	productHandler := handlers.NewProductHandler(productService, imageStorage)
	kitHandler             := handlers.NewKitHandler(kitService)
	orderHandler           := handlers.NewOrderHandler(orderService, cartService, paymentService)
	webhookHandler         := handlers.NewWebhookHandler(paymentService, orderService, cfg.MPWebhookSecret)
	distributorLeadHandler := handlers.NewDistributorLeadHandler(distributorLeadService)
	wizardHandler := handlers.NewWizardRecommendationHandler(wizardService)

	// ── Router ────────────────────────────────────────────────────────────────
	router := gin.New()
	router.Use(gin.Recovery())
	if cfg.GinMode != gin.TestMode {
		router.Use(gin.Logger())
	}
	router.Use(CorsMiddleware())

	router.Static("/uploads", "./uploads")
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── /api/v1 ───────────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")

	// ── Rutas PÚBLICAS — sin ningún middleware de autenticación ───────────────
	public := v1.Group("")
	{
		// Webhooks (MP llama sin JWT; la firma HMAC se valida internamente)
		public.POST("/webhooks/mercadopago", webhookHandler.HandleMercadoPago)

		// Auth
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/refresh", authHandler.Refresh)

		// Productos (lectura)
		public.GET("/products", productHandler.ListProducts)
		public.GET("/products/category/:category", productHandler.ListProductsByCategory)
		public.GET("/products/brand/:brand", productHandler.ListProductsByBrand)
		public.GET("/products/:id", productHandler.GetProduct)

		// Kits (lectura)
		public.GET("/kits", kitHandler.ListKits)
		public.GET("/kits/:id", kitHandler.GetKit)

		// Formularios de contacto / analytics
		public.POST("/distributor-leads", distributorLeadHandler.CreateLead)
		public.POST("/wizard-recommendations", wizardHandler.CreateRecommendation)
	}

	// ── Rutas PROTEGIDAS — requieren JWT de cualquier usuario autenticado ─────
	protected := v1.Group("")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)

		protected.POST("/checkout", orderHandler.Checkout)
		protected.GET("/orders/me", orderHandler.GetMyOrders)
		protected.GET("/orders/:id", orderHandler.GetOrder)
		protected.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)
	}

	// ── Rutas de ADMINISTRADOR — requieren JWT con role="admin" ───────────────
	admin := v1.Group("/admin")
	admin.Use(handlers.AuthMiddleware(authService), handlers.AdminMiddleware())
	{
		admin.GET("/users", userHandler.ListUsers)
		admin.DELETE("/users/:id", userHandler.DeleteUser)

		admin.GET("/products", productHandler.ListAdminProducts)
		admin.POST("/products", productHandler.CreateProduct)
		admin.PUT("/products/:id", productHandler.UpdateProduct)
		admin.PATCH("/products/:id/variants/:sku/stock", productHandler.UpdateVariantStock)
		admin.DELETE("/products/:id", productHandler.DeleteProduct)

		admin.POST("/kits", kitHandler.CreateKit)
		admin.DELETE("/kits/:id", kitHandler.DeleteKit)

		admin.GET("/orders", orderHandler.ListAllOrders)
		admin.PATCH("/orders/:id/status", orderHandler.UpdateAdminOrderStatus)
		admin.DELETE("/orders/:id", orderHandler.DeleteOrder)
	}

	return router
}

// CorsMiddleware permite peticiones cross-origin desde el frontend.
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
