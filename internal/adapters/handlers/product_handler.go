package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// maxMultipartMemory limita la memoria usada al parsear formularios multipart.
// El excedente se vuelca a archivos temporales — no es un límite duro de tamaño.
const maxMultipartMemory = 10 << 20 // 10 MiB

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

// CreateProduct maneja POST /api/products (solo admin).
// Espera multipart/form-data con campos: name, description, category, brand,
// base_price (string numérico), variants (JSON), specs (JSON), image (archivo opcional).
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("multipart inválido: %v", err)})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")
	brand := c.PostForm("brand")
	basePriceStr := c.PostForm("base_price")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'name' es requerido"})
		return
	}
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'category' es requerido"})
		return
	}
	if basePriceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'base_price' es requerido"})
		return
	}

	basePrice, err := strconv.ParseFloat(basePriceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("base_price inválido: %v", err)})
		return
	}
	if basePrice < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_price debe ser >= 0"})
		return
	}

	variants, err := parseVariantsForm(c.PostForm("variants"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("variants inválido: %v", err)})
		return
	}

	specs, err := parseSpecsForm(c.PostForm("specs"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("specs inválido: %v", err)})
		return
	}

	mainSpecs, err := parseMainSpecsForm(c.PostForm("main_specs"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("main_specs inválido: %v", err)})
		return
	}

	images, err := h.uploadImagesIfPresent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("image inválido: %v", err)})
		return
	}

	stock, err := parseStockForm(c.PostForm("stock"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("stock inválido: %v", err)})
		return
	}

	discountPercent, err := parseDiscountPercentForm(c.PostForm("discount_percent"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("discount_percent inválido: %v", err)})
		return
	}

	visibleTrue := true
	product := &domain.Product{
		Name:            name,
		Description:     description,
		BasePrice:       basePrice,
		DiscountPercent: discountPercent,
		Stock:           stock,
		Category:        category,
		Brand:           brand,
		Images:          images,
		Variants:        variants,
		Specs:           specs,
		MainSpecs:       mainSpecs,
		Visible:         &visibleTrue,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.productService.CreateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// parseVariantsForm decodifica el JSON stringificado del campo 'variants'.
// Cadena vacía → slice nil sin error (campo opcional).
func parseVariantsForm(raw string) ([]domain.Variant, error) {
	if raw == "" {
		return nil, nil
	}
	var variants []domain.Variant
	if err := json.Unmarshal([]byte(raw), &variants); err != nil {
		return nil, err
	}
	return variants, nil
}

// parseSpecsForm decodifica el JSON stringificado del campo 'specs'.
// Cadena vacía → slice nil sin error (campo opcional).
func parseSpecsForm(raw string) ([]domain.Spec, error) {
	if raw == "" {
		return nil, nil
	}
	var specs []domain.Spec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, err
	}
	return specs, nil
}

// parseMainSpecsForm decodifica el JSON stringificado del campo 'main_specs'.
func parseMainSpecsForm(raw string) ([]domain.MainSpec, error) {
	if raw == "" {
		return nil, nil
	}
	var ms []domain.MainSpec
	if err := json.Unmarshal([]byte(raw), &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

// parseStockForm convierte el valor textual del formulario en entero >= 0.
// Acepta enteros y floats (el frontend normaliza "12,5" → "12.5"); el float se trunca.
// Cadena vacía → 0 sin error (campo opcional).
func parseStockForm(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("debe ser un número: %v", err)
	}
	if f < 0 {
		return 0, errors.New("debe ser >= 0")
	}
	return int(f), nil
}

// parseDiscountPercentForm parsea el campo 'discount_percent' como número en [0, 100].
// Cadena vacía → 0 sin error (campo opcional).
func parseDiscountPercentForm(raw string) (float64, error) {
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("debe ser un número: %v", err)
	}
	if v < 0 || v > 100 {
		return 0, errors.New("debe estar entre 0 y 100")
	}
	return v, nil
}

// uploadImagesIfPresent sube todos los archivos con key "image" del formulario.
// Devuelve slice nil sin error cuando no se adjuntó ningún archivo.
func (h *ProductHandler) uploadImagesIfPresent(c *gin.Context) ([]string, error) {
	if c.Request.MultipartForm == nil {
		return nil, nil
	}
	files := c.Request.MultipartForm.File["image"]
	if len(files) == 0 {
		return nil, nil
	}
	var urls []string
	for _, fh := range files {
		url, err := h.imageStorage.UploadImage(c.Request.Context(), fh)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
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

	// Exclude products explicitly hidden (visible=false). Missing/nil/true → visible.
	visibilityFilter := map[string]interface{}{
		"visible": map[string]interface{}{"$ne": false},
	}
	products, err := h.productService.ListProducts(c.Request.Context(), visibilityFilter, skip, limit)
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

// UpdateProduct maneja PUT /api/products/:id (solo admin).
// Espera multipart/form-data. Semántica parcial: solo se actualizan los campos presentes.
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("multipart inválido: %v", err)})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	form := c.Request.PostForm

	if form.Has("name") {
		v := form.Get("name")
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'name' no puede ser vacío"})
			return
		}
		product.Name = v
	}
	if form.Has("description") {
		product.Description = form.Get("description")
	}
	if form.Has("category") {
		v := form.Get("category")
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'category' no puede ser vacío"})
			return
		}
		product.Category = v
	}
	if form.Has("brand") {
		product.Brand = form.Get("brand")
	}
	if form.Has("base_price") {
		basePrice, err := strconv.ParseFloat(form.Get("base_price"), 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("base_price inválido: %v", err)})
			return
		}
		if basePrice < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "base_price debe ser >= 0"})
			return
		}
		product.BasePrice = basePrice
	}
	if form.Has("stock") {
		stock, err := parseStockForm(form.Get("stock"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("stock inválido: %v", err)})
			return
		}
		product.Stock = stock
	}
	if form.Has("discount_percent") {
		discount, err := parseDiscountPercentForm(form.Get("discount_percent"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("discount_percent inválido: %v", err)})
			return
		}
		product.DiscountPercent = discount
	}
	if form.Has("variants") {
		variants, err := parseVariantsForm(form.Get("variants"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("variants inválido: %v", err)})
			return
		}
		product.Variants = variants
	}
	if form.Has("specs") {
		specs, err := parseSpecsForm(form.Get("specs"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("specs inválido: %v", err)})
			return
		}
		product.Specs = specs
	}
	if form.Has("main_specs") {
		ms, err := parseMainSpecsForm(form.Get("main_specs"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("main_specs inválido: %v", err)})
			return
		}
		product.MainSpecs = ms
	}
	if form.Has("images") {
		var imgs []string
		if err := json.Unmarshal([]byte(form.Get("images")), &imgs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("images inválido: %v", err)})
			return
		}
		product.Images = imgs
	}
	newImages, err := h.uploadImagesIfPresent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("image inválido: %v", err)})
		return
	}
	if len(newImages) > 0 {
		product.Images = append(product.Images, newImages...)
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

// SetVisibility maneja PATCH /admin/products/:id/visibility (solo admin)
func (h *ProductHandler) SetVisibility(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Visible bool `json:"visible"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	product.Visible = &req.Visible
	product.UpdatedAt = time.Now()

	if err := h.productService.UpdateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// SetSortOrder maneja PATCH /admin/products/:id/sort-order (solo admin)
func (h *ProductHandler) SetSortOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		SortOrder int `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	product.SortOrder = req.SortOrder
	product.UpdatedAt = time.Now()

	if err := h.productService.UpdateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
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
