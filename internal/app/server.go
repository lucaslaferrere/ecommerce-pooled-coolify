// Package app wires the full dependency graph and returns a configured Gin router.
// Both cmd/api/main.go and the integration tests use BuildRouter so that tests
// exercise the exact same middleware stack and route table as production.
package app

import (
	"log"
	"net/http"
	"strings"
	"time"

	"ecommerce-pooled/internal/adapters/handlers"
	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/config"
	"ecommerce-pooled/internal/core/services"
	"ecommerce-pooled/internal/realtime"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// BuildRouter wires repositories → services → handlers and registers all routes
// under /api/v1.  gin.Logger() is only added when GinMode != "test" so that
// integration test output stays clean.
func BuildRouter(cfg *config.Config, db *mongo.Database) *gin.Engine {
	authLimiter := handlers.NewRateLimiter(10, time.Minute) // 10 intentos por minuto por IP
	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo             := repositories.NewUserRepositoryMongo(db.Collection("users"))
	productRepo          := repositories.NewProductRepositoryMongo(db.Collection("products"))
	passwordResetRepo    := repositories.NewPasswordResetRepositoryMongo(db.Collection("password_reset_tokens"))
	kitRepo             := repositories.NewKitRepositoryMongo(db.Collection("kits"))
	orderRepo           := repositories.NewOrderRepositoryMongo(db.Collection("orders"))
	distributorLeadRepo := repositories.NewDistributorLeadRepositoryMongo(db.Collection("distributor_leads"))
	wizardRepo := repositories.NewWizardRecommendationRepositoryMongo(db.Collection("wizard_recommendations"))
	eventRepo      := repositories.NewEventRepositoryMongo(db.Collection("events"))
	cartLinkRepo   := repositories.NewCartLinkRepositoryMongo(db.Collection("cart_links"))
	couponRepo     := repositories.NewCouponRepositoryMongo(db.Collection("coupons"))
	settingRepo    := repositories.NewSettingRepositoryMongo(db.Collection("settings"))
	cartRepo       := repositories.NewCartRepositoryMongo(db.Collection("carts"))
	invoiceRepo    := repositories.NewInvoiceRepositoryMongo(db.Collection("invoices"))
	emailTemplateRepo := repositories.NewEmailTemplateRepositoryMongo(db.Collection("email_templates"))

	// ── Services ─────────────────────────────────────────────────────────────
	realtimeHub := realtime.NewHub()
	emailService := services.NewEmailService(cfg.ResendAPIKey, cfg.ResendFrom, cfg.OwnerEmails)
	log.Printf("[BOOT] owner emails configurados: %d → %v", len(cfg.OwnerEmails), cfg.OwnerEmails)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret).
		WithPasswordReset(passwordResetRepo, emailService)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	kitService             := services.NewKitService(kitRepo)
	cartService            := services.NewCartService(productRepo, kitRepo)
	orderService           := services.NewOrderService(orderRepo, productRepo, realtimeHub)
	distributorLeadService := services.NewDistributorLeadService(distributorLeadRepo)
	wizardService := services.NewWizardRecommendationService(wizardRepo)
	eventService     := services.NewEventService(eventRepo, orderRepo)
	geoService       := services.NewGeoService(cfg.GeoIPDBPath)
	activeSessions   := realtime.NewActiveSessions()
	cartLinkService  := services.NewCartLinkService(cartLinkRepo)
	couponService    := services.NewCouponService(couponRepo, orderRepo)
	settingService := services.NewSettingService(settingRepo)
	emailTemplateService := services.NewEmailTemplateService(emailTemplateRepo, settingService)
	cartStorageService := services.NewCartStorageService(cartRepo)
	invoiceService := services.NewInvoiceService(invoiceRepo, orderService, emailService)
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
	authHandler := handlers.NewAuthHandler(authService, emailService)
	userHandler := handlers.NewUserHandler(userService)
	productHandler := handlers.NewProductHandler(productService, imageStorage)
	kitHandler             := handlers.NewKitHandler(kitService, imageStorage)
	couponHandler          := handlers.NewCouponHandler(couponService)
	settingHandler := handlers.NewSettingHandler(settingService)
	orderHandler           := handlers.NewOrderHandler(orderService, cartService, paymentService, emailService, couponService, cartStorageService)
	webhookHandler         := handlers.NewWebhookHandler(paymentService, orderService, emailService, cfg.MPWebhookSecret)
	distributorLeadHandler := handlers.NewDistributorLeadHandler(distributorLeadService)
	wizardHandler   := handlers.NewWizardRecommendationHandler(wizardService)
	eventHandler    := handlers.NewEventHandler(eventService, geoService, activeSessions, realtimeHub)
	realtimeHandler := handlers.NewRealtimeHandler(realtimeHub, eventService, activeSessions)
	warrantyHandler  := handlers.NewWarrantyHandler(cfg.ResendAPIKey, cfg.ResendFrom)
	cartLinkHandler      := handlers.NewCartLinkHandler(cartLinkService, cfg.FrontendURL)
	cartStorageHandler   := handlers.NewCartStorageHandler(cartStorageService, couponService, emailService, userService, emailTemplateService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
	emailTemplateHandler := handlers.NewEmailTemplateHandler(emailTemplateService)

	log.Println("[BOOT] server.go v2 — cart-links registrado")

	// ── Router ────────────────────────────────────────────────────────────────
	router := gin.New()
	router.Use(gin.Recovery())
	if cfg.GinMode != gin.TestMode {
		router.Use(gin.Logger())
	}
	router.Use(CorsMiddleware(cfg.AllowedOrigin))

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

		// Auth (rate limited: 10 req/min por IP)
		public.POST("/auth/register", authLimiter.Middleware(), authHandler.Register)
		public.POST("/auth/login", authLimiter.Middleware(), authHandler.Login)
		public.POST("/auth/refresh", authLimiter.Middleware(), authHandler.Refresh)
		public.POST("/auth/forgot-password", authLimiter.Middleware(), authHandler.ForgotPassword)
		public.POST("/auth/reset-password", authLimiter.Middleware(), authHandler.ResetPassword)

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

		// Eventos de analytics (fire-and-forget, sin auth)
		public.POST("/events", eventHandler.TrackEvent)

		// Garantía
		public.POST("/warranty", warrantyHandler.Submit)

		// Cart links (lectura pública para que el cliente cargue el carrito)
		public.GET("/cart-links/:token", cartLinkHandler.Get)

		// Cupones (validación pública desde el checkout)
		public.POST("/coupons/validate", couponHandler.Validate)
	}

	// ── Rutas PROTEGIDAS — requieren JWT de cualquier usuario autenticado ─────
	protected := v1.Group("")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)

		protected.PUT("/cart", cartStorageHandler.SaveMyCart)
		protected.DELETE("/cart", cartStorageHandler.ClearMyCart)

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

		// Cart links (solo admin puede crear)
		admin.POST("/cart-links", cartLinkHandler.Create)

		admin.GET("/products", productHandler.ListAdminProducts)
		admin.POST("/products", productHandler.CreateProduct)
		admin.PUT("/products/:id", productHandler.UpdateProduct)
		admin.PATCH("/products/bulk-price", productHandler.BulkUpdatePrice)
		admin.PATCH("/products/:id/visibility", productHandler.SetVisibility)
		admin.PATCH("/products/:id/sort-order", productHandler.SetSortOrder)
		admin.PATCH("/products/:id/variants/:sku/stock", productHandler.UpdateVariantStock)
		admin.DELETE("/products/:id", productHandler.DeleteProduct)

		admin.GET("/kits", kitHandler.ListAllKits)
		admin.POST("/kits", kitHandler.CreateKit)
		admin.PUT("/kits/:id", kitHandler.UpdateKit)
		admin.PATCH("/kits/:id/visibility", kitHandler.SetKitVisibility)
		admin.PATCH("/kits/:id/sort-order", kitHandler.SetKitSortOrder)
		admin.PATCH("/kits/bulk-price", kitHandler.BulkUpdatePrice)
		admin.DELETE("/kits/:id", kitHandler.DeleteKit)

		admin.GET("/coupons", couponHandler.List)
		admin.POST("/coupons", couponHandler.Create)
		admin.PUT("/coupons/:id", couponHandler.Update)
		admin.DELETE("/coupons/:id", couponHandler.Delete)

		admin.GET("/analytics", eventHandler.GetAnalytics)
		admin.GET("/analytics/sales", eventHandler.GetSalesSeries)
		admin.GET("/analytics/overview", eventHandler.GetSalesOverview)
		admin.GET("/traffic", eventHandler.GetTraffic)
		admin.GET("/realtime/stream", realtimeHandler.Stream)
		admin.GET("/realtime/snapshot", realtimeHandler.Snapshot)

		admin.GET("/orders", orderHandler.ListAllOrders)
		admin.PATCH("/orders/:id/status", orderHandler.UpdateAdminOrderStatus)
		admin.DELETE("/orders/:id", orderHandler.DeleteOrder)
		admin.POST("/orders/:id/invoice", invoiceHandler.Upload)
		admin.POST("/orders/:id/invoice/resend", invoiceHandler.Resend)

		admin.GET("/settings/abandoned-cart-email", settingHandler.GetAbandonedCartEmail)
		admin.PUT("/settings/abandoned-cart-email", settingHandler.SetAbandonedCartEmail)

		admin.GET("/carts", cartStorageHandler.ListCarts)
		admin.POST("/carts/:userID/send-coupon", cartStorageHandler.SendCoupon)

		admin.GET("/email-templates", emailTemplateHandler.List)
		admin.POST("/email-templates", emailTemplateHandler.Create)
		admin.DELETE("/email-templates/:id", emailTemplateHandler.Delete)
	}

	return router
}

// CorsMiddleware permite peticiones cross-origin solo desde el origen configurado.
// En desarrollo también acepta localhost en cualquier puerto.
func CorsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin == allowedOrigin ||
			strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:")

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}
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
