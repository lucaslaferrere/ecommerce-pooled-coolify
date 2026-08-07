package handlers

import (
	"net/http"

	"ecommerce-pooled/internal/core/services"

	"github.com/gin-gonic/gin"
)

// UploadHandler expone subida de archivos sueltos, sin asociarlos a ninguna
// entidad. Se usa para asignarle a una variante una foto propia, independiente
// de la galería general del producto.
type UploadHandler struct {
	imageStorage services.ImageStorageService
}

func NewUploadHandler(imageStorage services.ImageStorageService) *UploadHandler {
	return &UploadHandler{imageStorage: imageStorage}
}

// UploadImage maneja POST /api/v1/admin/uploads/image (solo admin).
// Sube un único archivo (campo "image") y devuelve su URL.
func (h *UploadHandler) UploadImage(c *gin.Context) {
	fh, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'image' (archivo) requerido"})
		return
	}
	url, err := h.imageStorage.UploadImage(c.Request.Context(), fh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}
