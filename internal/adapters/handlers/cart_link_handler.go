package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

type CartLinkHandler struct {
	service *services.CartLinkService
	appURL  string
}

func NewCartLinkHandler(service *services.CartLinkService, appURL string) *CartLinkHandler {
	return &CartLinkHandler{service: service, appURL: appURL}
}

type createCartLinkRequest struct {
	Items []domain.CartLinkItem `json:"items" binding:"required,min=1"`
}

// POST /api/v1/admin/cart-links
func (h *CartLinkHandler) Create(c *gin.Context) {
	var req createCartLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.service.Create(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	frontendURL := h.appURL + "/checkout?cart=" + link.Token
	c.JSON(http.StatusCreated, gin.H{
		"token":      link.Token,
		"url":        frontendURL,
		"expires_at": link.ExpiresAt,
	})
}

// GET /api/v1/cart-links/:token
func (h *CartLinkHandler) Get(c *gin.Context) {
	token := c.Param("token")
	link, err := h.service.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, link)
}
