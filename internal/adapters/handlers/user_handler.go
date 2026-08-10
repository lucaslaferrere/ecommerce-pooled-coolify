package handlers

import (
	"net/http"
	"strconv"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler maneja las peticiones HTTP relacionadas con usuarios
type UserHandler struct {
	userService *services.UserService
	authService *services.AuthService
}

// NewUserHandler crea una nueva instancia de UserHandler
func NewUserHandler(userService *services.UserService, authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
	}
}

// CreateUser maneja POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required,oneof=admin client"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         req.Role,
	}

	if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser maneja GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Only the owner or an admin can fetch a user profile.
	if callerRole.(string) != "admin" && callerID.(string) != idStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	user.PasswordHash = "" // never expose the hash
	c.JSON(http.StatusOK, user)
}

// GetUserByEmail maneja GET /api/users/email/:email
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers maneja GET /api/users con paginación
func (h *UserHandler) ListUsers(c *gin.Context) {
	skip, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	users, err := h.userService.ListUsers(c.Request.Context(), skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Nunca exponer el hash de contraseña en el listado.
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// UpdateUser maneja PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	callerID, _ := c.Get("user_id")
	callerRole, _ := c.Get("role")

	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Only the owner or an admin can update a user profile.
	if callerRole.(string) != "admin" && callerID.(string) != idStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing user to preserve role and other fields.
	existing, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil || existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	if req.Email != "" {
		existing.Email = req.Email
	}
	// Role is intentionally not updatable through this endpoint.

	if err := h.userService.UpdateUser(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existing.PasswordHash = ""
	c.JSON(http.StatusOK, existing)
}

// DeleteUser maneja DELETE /api/users/:id
// Nadie puede borrarse a sí mismo. Para borrar un admin/superadmin, quien
// pide el borrado tiene que ser superadmin (a un cliente lo puede borrar
// cualquier admin, como hasta ahora).
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	actorIDRaw, _ := c.Get("user_id")
	if actorIDStr, ok := actorIDRaw.(string); ok && actorIDStr == idStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "no podés eliminar tu propia cuenta"})
		return
	}

	target, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if target != nil && (target.Role == "admin" || target.Role == "superadmin") {
		actorRole, _ := c.Get("role")
		if role, _ := actorRole.(string); role != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "solo un superadmin puede eliminar a un administrador"})
			return
		}
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "usuario eliminado"})
}

// RegisterRoutes registra las rutas de usuario (DEPRECATED - usar main.go)
func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	users := router.Group("/api/users")
	{
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUser)
		users.GET("/email/:email", h.GetUserByEmail)
		users.GET("", h.ListUsers)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}
}

// InviteAdmin maneja POST /admin/users/invite-admin (solo superadmin).
// Da de alta (o promueve) a un usuario como admin y le manda un mail para
// que defina su propia contraseña.
func (h *UserHandler) InviteAdmin(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.InviteAdmin(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"role":  user.Role,
	})
}

// ChangeUserRole maneja PATCH /admin/users/:id/role (solo superadmin).
// Promueve o degrada entre "client" y "admin". No se puede usar sobre uno
// mismo, ni para tocar a un superadmin.
func (h *UserHandler) ChangeUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	actorIDRaw, _ := c.Get("user_id")
	if actorIDStr, ok := actorIDRaw.(string); ok && actorIDStr == idStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "no podés cambiar tu propio rol"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=admin client"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.ChangeUserRole(c.Request.Context(), id, req.Role)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"role":  user.Role,
	})
}
