package handlers

import (
	"log"
	"net/http"
	"strings"

	"ecommerce-pooled/internal/core/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler maneja las peticiones HTTP relacionadas con autenticación
type AuthHandler struct {
	authService  *services.AuthService
	emailService *services.EmailService
}

// NewAuthHandler crea una nueva instancia de AuthHandler
func NewAuthHandler(authService *services.AuthService, emailService *services.EmailService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
	}
}

// RegisterRequest contiene los datos para registrar un nuevo usuario.
// Role es opcional: si se omite el frontend recibirá role="client" por defecto;
// si se envía debe ser "admin" o "client".
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"    binding:"omitempty,oneof=admin client"`
}

// LoginRequest contiene las credenciales para iniciar sesión
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest contiene el refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register maneja POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Always register as client — role cannot be set by the caller.
	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, "client")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Notificar a los dueños del nuevo registro (fire-and-forget)
	if h.emailService != nil {
		go h.emailService.SendOwnerNewUser(user)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID.Hex(),
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// Login maneja POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Refresh maneja POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// ForgotPasswordRequest contiene el email para solicitar recuperación.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest contiene los datos para restablecer la contraseña.
type ResetPasswordRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Code        string `json:"code"         binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ForgotPassword maneja POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.authService.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		log.Printf("forgot-password: %v", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Si el email está registrado recibirás un código en breve"})
}

// ResetPassword maneja POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada correctamente"})
}

// RegisterRoutes registra las rutas de autenticación
func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}
}

// ============================================================================
// MIDDLEWARES DE AUTENTICACIÓN Y AUTORIZACIÓN
// ============================================================================

// AuthMiddleware valida el JWT y extrae la información del usuario
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			c.Abort()
			return
		}

		// Extraer el token (formato: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "formato de token inválido"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validar token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			c.Abort()
			return
		}

		// Guardar información en contexto
		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("role", claims["role"])
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireAuth es un middleware que requiere autenticación
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("user_id"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminMiddleware verifica que el usuario tenga rol "admin" o "superadmin"
// (superadmin es superconjunto: puede todo lo que puede un admin, y además
// gestionar otros admins vía SuperAdminMiddleware en rutas específicas).
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok || (userRole != "admin" && userRole != "superadmin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado: se requiere rol admin"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminMiddleware verifica que el usuario tenga rol "superadmin". Se usa
// además de AdminMiddleware, en rutas puntuales de gestión de administradores
// (invitar, promover, degradar, eliminar).
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok || userRole != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado: se requiere rol superadmin"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ClientMiddleware verifica que el usuario tenga rol "client"
func ClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "autenticación requerida"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok || userRole != "client" {
			c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado: se requiere rol client"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware es un middleware de autenticación opcional
// Si el token es válido, extrae la información; si no, continúa sin ella
func OptionalAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			// Token inválido, pero continuamos sin error
			c.Next()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("role", claims["role"])
		c.Set("claims", claims)

		c.Next()
	}
}
