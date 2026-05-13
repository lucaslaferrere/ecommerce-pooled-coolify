package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KitHandler struct {
	kitService   *services.KitService
	imageStorage services.ImageStorageService
}

func NewKitHandler(kitService *services.KitService, imageStorage services.ImageStorageService) *KitHandler {
	return &KitHandler{kitService: kitService, imageStorage: imageStorage}
}

// ListKits maneja GET /api/v1/kits
func (h *KitHandler) ListKits(c *gin.Context) {
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	if c.Query("featured") == "true" {
		kits, err := h.kitService.ListFeaturedKits(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, kits)
		return
	}

	kits, err := h.kitService.ListKits(c.Request.Context(), skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kits)
}

// GetKit maneja GET /api/v1/kits/:id
func (h *KitHandler) GetKit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	kit, err := h.kitService.GetKit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kit)
}

// CreateKit maneja POST /api/v1/admin/kits (multipart/form-data)
func (h *KitHandler) CreateKit(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart inválido"})
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'name' es requerido"})
		return
	}

	priceStr := c.PostForm("price")
	if priceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'price' es requerido"})
		return
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price inválido: debe ser > 0"})
		return
	}

	var originalPrice *float64
	if v := c.PostForm("original_price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "original_price inválido"})
			return
		}
		originalPrice = &f
	}

	productIDs, err := parseProductIDsForm(c.PostForm("product_ids"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_ids inválido: " + err.Error()})
		return
	}

	featured := strings.ToLower(c.PostForm("featured")) == "true"

	slug := slugify(name)

	imageURL, err := h.uploadKitImageIfPresent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image inválido: " + err.Error()})
		return
	}

	kit := &domain.Kit{
		Name:          name,
		Slug:          slug,
		Description:   c.PostForm("description"),
		Price:         price,
		OriginalPrice: originalPrice,
		ImageURL:      imageURL,
		PoolSize:      c.PostForm("pool_size"),
		ProductIDs:    productIDs,
		Featured:      featured,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.kitService.CreateKit(c.Request.Context(), kit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, kit)
}

// UpdateKit maneja PUT /api/v1/admin/kits/:id (multipart/form-data, actualización parcial)
func (h *KitHandler) UpdateKit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart inválido"})
		return
	}

	kit, err := h.kitService.GetKit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	form := c.Request.PostForm

	if form.Has("name") {
		v := form.Get("name")
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name no puede ser vacío"})
			return
		}
		kit.Name = v
		kit.Slug = slugify(v)
	}
	if form.Has("description") {
		kit.Description = form.Get("description")
	}
	if form.Has("price") {
		p, err := strconv.ParseFloat(form.Get("price"), 64)
		if err != nil || p <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price inválido"})
			return
		}
		kit.Price = p
	}
	if form.Has("original_price") {
		v := form.Get("original_price")
		if v == "" {
			kit.OriginalPrice = nil
		} else {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "original_price inválido"})
				return
			}
			kit.OriginalPrice = &f
		}
	}
	if form.Has("pool_size") {
		kit.PoolSize = form.Get("pool_size")
	}
	if form.Has("featured") {
		kit.Featured = strings.ToLower(form.Get("featured")) == "true"
	}
	if form.Has("product_ids") {
		ids, err := parseProductIDsForm(form.Get("product_ids"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_ids inválido"})
			return
		}
		kit.ProductIDs = ids
	}

	newImageURL, err := h.uploadKitImageIfPresent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image inválido: " + err.Error()})
		return
	}
	if newImageURL != "" {
		kit.ImageURL = newImageURL
	}

	kit.UpdatedAt = time.Now()

	if err := h.kitService.UpdateKit(c.Request.Context(), kit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kit)
}

// DeleteKit maneja DELETE /api/v1/admin/kits/:id
func (h *KitHandler) DeleteKit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.kitService.DeleteKit(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kit eliminado"})
}

func (h *KitHandler) uploadKitImageIfPresent(c *gin.Context) (string, error) {
	fh, err := c.FormFile("image")
	if err != nil {
		return "", nil
	}
	url, err := h.imageStorage.UploadImage(c.Request.Context(), fh)
	if err != nil {
		return "", err
	}
	return url, nil
}

func parseProductIDsForm(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// slugify genera un slug simple a partir del nombre (minúsculas, espacios → guiones).
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
