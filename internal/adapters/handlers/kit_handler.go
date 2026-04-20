package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KitHandler struct {
	kitService *services.KitService
}

func NewKitHandler(kitService *services.KitService) *KitHandler {
	return &KitHandler{kitService: kitService}
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
		c.JSON(http.StatusOK, gin.H{"kits": kits, "total": len(kits)})
		return
	}

	kits, err := h.kitService.ListKits(c.Request.Context(), skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"kits": kits, "total": len(kits)})
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

// CreateKit maneja POST /api/v1/admin/kits
func (h *KitHandler) CreateKit(c *gin.Context) {
	var req struct {
		Name          string   `json:"name" binding:"required"`
		Slug          string   `json:"slug" binding:"required"`
		Description   string   `json:"description"`
		Price         float64  `json:"price" binding:"required,gt=0"`
		OriginalPrice *float64 `json:"original_price"`
		ImageURL      string   `json:"image_url"`
		PoolSize      string   `json:"pool_size"`
		ProductIDs    []string `json:"product_ids"`
		Featured      bool     `json:"featured"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kit := &domain.Kit{
		Name:          req.Name,
		Slug:          req.Slug,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		ImageURL:      req.ImageURL,
		PoolSize:      req.PoolSize,
		ProductIDs:    req.ProductIDs,
		Featured:      req.Featured,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.kitService.CreateKit(c.Request.Context(), kit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, kit)
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
