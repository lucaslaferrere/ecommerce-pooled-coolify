package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductHandler maneja las peticiones HTTP relacionadas con productos
type ProductHandler struct {
	productService *services.ProductService
	imageStorage   services.ImageStorageService
}

// NewProductHandler crea una nueva instancia de ProductHandler
func NewProductHandler(productService *services.ProductService, imageStorage services.ImageStorageService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		imageStorage:   imageStorage,
	}
}

// CreateProduct maneja POST /api/products (solo admin)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")
	brand := c.PostForm("brand")
	basePriceStr := c.PostForm("base_price")
	variantsJSON := c.PostForm("variants")

	if name == "" || category == "" || basePriceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faltan campos requeridos: name, category, base_price"})
		return
	}

	basePrice, err := strconv.ParseFloat(basePriceStr, 64)
	if err != nil || basePrice < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_price inválido"})
		return
	}

	var variants []domain.Variant
	if variantsJSON != "" {
		if err := json.Unmarshal([]byte(variantsJSON), &variants); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "variants inválido"})
			return
		}
	}

	// Manejar imagen si viene
	var images []string
	file, err := c.FormFile("image")
	if err == nil {
		savedURL, err := h.imageStorage.UploadImage(c.Request.Context(), file)
		if err == nil {
			images = append(images, savedURL)
		}
	}

	product := &domain.Product{
		Name:        name,
		Description: description,
		BasePrice:   basePrice,
		Category:    category,
		Brand:       brand,
		Images:      images,
		Variants:    variants,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.productService.CreateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct maneja GET /api/products/:id (público)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListAdminProducts maneja GET /api/v1/admin/products (solo admin)
func (h *ProductHandler) ListAdminProducts(c *gin.Context) {
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)

	products, err := h.productService.ListProducts(c.Request.Context(), map[string]interface{}{}, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products) // array directo, no wrapped
}

// ListProducts maneja GET /api/products (público)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	products, err := h.productService.ListProducts(c.Request.Context(), map[string]interface{}{}, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    len(products),
		"skip":     skip,
		"limit":    limit,
	})
}

// ListProductsByCategory maneja GET /api/products/category/:category (público)
func (h *ProductHandler) ListProductsByCategory(c *gin.Context) {
	category := c.Param("category")

	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	products, err := h.productService.GetProductsByCategory(c.Request.Context(), category, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    len(products),
		"category": category,
	})
}

// ListProductsByBrand maneja GET /api/products/brand/:brand (público)
func (h *ProductHandler) ListProductsByBrand(c *gin.Context) {
	brand := c.Param("brand")

	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	products, err := h.productService.GetProductsByBrand(c.Request.Context(), brand, skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    len(products),
		"brand":    brand,
	})
}

// UpdateProduct maneja PUT /api/products/:id (solo admin)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Name        string           `json:"name"`
		Description string           `json:"description"`
		BasePrice   float64          `json:"base_price"`
		Category    string           `json:"category"`
		Brand       string           `json:"brand"`
		Images      []string         `json:"images"`
		Variants    []domain.Variant `json:"variants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener producto existente
	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Actualizar campos
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.BasePrice > 0 {
		product.BasePrice = req.BasePrice
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Brand != "" {
		product.Brand = req.Brand
	}
	if len(req.Images) > 0 {
		product.Images = req.Images
	}
	if len(req.Variants) > 0 {
		product.Variants = req.Variants
	}

	product.UpdatedAt = time.Now()

	if err := h.productService.UpdateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateVariantStock maneja PUT /api/products/:id/variants/:sku/stock (solo admin)
func (h *ProductHandler) UpdateVariantStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU requerido"})
		return
	}

	var req struct {
		Stock int `json:"stock" binding:"required,gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productService.ReduceVariantStock(c.Request.Context(), id, sku, req.Stock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock actualizado"})
}

// DeleteProduct maneja DELETE /api/products/:id (solo admin)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "producto eliminado"})
}

// RegisterRoutes registra las rutas de producto
// Nota: Las rutas de mutación (POST, PUT, DELETE) deben incluir AdminMiddleware en main.go
func (h *ProductHandler) RegisterRoutes(router *gin.Engine) {
	products := router.Group("/api/products")
	{
		// Rutas públicas
		products.GET("", h.ListProducts)
		products.GET("/:id", h.GetProduct)
		products.GET("/category/:category", h.ListProductsByCategory)
		products.GET("/brand/:brand", h.ListProductsByBrand)

		// Rutas de admin (se aplican en main.go)
		// products.POST("", AdminMiddleware(), h.CreateProduct)
		// products.PUT("/:id", AdminMiddleware(), h.UpdateProduct)
		// products.PUT("/:id/variants/:sku/stock", AdminMiddleware(), h.UpdateVariantStock)
		// products.DELETE("/:id", AdminMiddleware(), h.DeleteProduct)
	}
}
