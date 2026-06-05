package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"ecommerce-pooled/internal/core/domain"
	"ecommerce-pooled/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenPair contiene los tokens de acceso y refresco
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AuthService proporciona servicios de autenticación
type AuthService struct {
	userRepository    ports.UserRepository
	passwordResetRepo ports.PasswordResetRepository
	emailService      *EmailService
	jwtSecret         string
}

// NewAuthService crea una nueva instancia de AuthService
func NewAuthService(userRepository ports.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
	}
}

// WithPasswordReset agrega soporte de recuperación de contraseña al AuthService.
func (s *AuthService) WithPasswordReset(repo ports.PasswordResetRepository, email *EmailService) *AuthService {
	s.passwordResetRepo = repo
	s.emailService = email
	return s
}

// Register registra un nuevo usuario
func (s *AuthService) Register(ctx context.Context, email string, password string, role string) (*domain.User, error) {
	// Validaciones
	if email == "" {
		return nil, errors.New("email requerido")
	}

	if password == "" {
		return nil, errors.New("contraseña requerida")
	}

	if len(password) < 6 {
		return nil, errors.New("la contraseña debe tener al menos 6 caracteres")
	}

	if role != "admin" && role != "client" {
		return nil, errors.New("rol debe ser 'admin' o 'client'")
	}

	// Verificar que el email no exista
	existingUser, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("el email ya está registrado")
	}

	// Hash de la contraseña con bcrypt (costo 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error al hashear contraseña: %w", err)
	}

	// Crear nuevo usuario
	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Guardar en base de datos
	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login verifica las credenciales y retorna tokens JWT
func (s *AuthService) Login(ctx context.Context, email string, password string) (*TokenPair, error) {
	// Validaciones
	if email == "" {
		return nil, errors.New("email requerido")
	}

	if password == "" {
		return nil, errors.New("contraseña requerida")
	}

	// Obtener usuario por email
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("credenciales inválidas")
	}

	// Verificar contraseña
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	// Generar tokens JWT
	tokens, err := s.GenerateTokens(user.ID.Hex(), user.Role, user.Email)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// GenerateTokens genera un par de JWT (Access y Refresh)
func (s *AuthService) GenerateTokens(userID string, role string, email string) (*TokenPair, error) {
	// Access Token - válido por 15 minutos
	accessTokenExpiry := time.Now().Add(24 * time.Hour)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"type":    "access",
		"exp":     accessTokenExpiry.Unix(),
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("error generando access token: %w", err)
	}

	// Refresh Token - válido por 7 días
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"type":    "refresh",
		"exp":     refreshTokenExpiry.Unix(),
		"iat":     time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("error generando refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    accessTokenExpiry.Unix() - time.Now().Unix(),
	}, nil
}

// ValidateToken valida un JWT y retorna los claims
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error validando token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	// Verificar que sea un access token
	if tokenType, ok := claims["type"]; !ok || tokenType != "access" {
		return nil, errors.New("tipo de token inválido")
	}

	return claims, nil
}

// RefreshToken genera un nuevo access token usando un refresh token
func (s *AuthService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	token, err := jwt.ParseWithClaims(refreshTokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error validando refresh token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("refresh token inválido")
	}

	// Verificar que sea un refresh token
	if tokenType, ok := claims["type"]; !ok || tokenType != "refresh" {
		return nil, errors.New("tipo de token inválido")
	}

	// Extraer información
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id no válido en token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("role no válido en token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("email no válido en token")
	}

	// Generar nuevos tokens
	return s.GenerateTokens(userID, role, email)
}

// ForgotPassword genera un código de verificación y lo envía por email.
// Siempre responde OK aunque el email no exista (evita enumeración de usuarios).
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	if s.passwordResetRepo == nil {
		return errors.New("servicio de recuperación no configurado")
	}
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("forgot-password: error buscando usuario: %v", err)
		return nil
	}
	if user == nil {
		log.Printf("forgot-password: usuario no encontrado para %s", email)
		return nil
	}
	log.Printf("forgot-password: enviando código a %s", email)
	if err := s.passwordResetRepo.InvalidateByEmail(ctx, email); err != nil {
		return err
	}
	code, err := generateNumericCode(6)
	if err != nil {
		return fmt.Errorf("error generando código: %w", err)
	}
	token := &domain.PasswordResetToken{
		UserID:    user.ID,
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}
	if err := s.passwordResetRepo.Create(ctx, token); err != nil {
		return err
	}
	return s.emailService.SendPasswordResetCode(email, code)
}

// ResetPassword verifica el código y actualiza la contraseña del usuario.
func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if s.passwordResetRepo == nil {
		return errors.New("servicio de recuperación no configurado")
	}
	if len(newPassword) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	resetToken, err := s.passwordResetRepo.GetActiveByEmailAndCode(ctx, email, code)
	if err != nil {
		return err
	}
	if resetToken == nil {
		return errors.New("código inválido o vencido")
	}
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("usuario no encontrado")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error al hashear contraseña: %w", err)
	}
	user.PasswordHash = string(hashed)
	user.UpdatedAt = time.Now()
	if err := s.userRepository.Update(ctx, user); err != nil {
		return err
	}
	return s.passwordResetRepo.MarkAsUsed(ctx, resetToken.ID)
}

func generateNumericCode(n int) (string, error) {
	const digits = "0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = digits[int(buf[i])%len(digits)]
	}
	return string(buf), nil
}

// GetUserFromToken extrae la información del usuario del JWT
func (s *AuthService) GetUserFromToken(claims jwt.MapClaims) (userID string, email string, role string, err error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", "", errors.New("user_id no válido")
	}

	email, ok = claims["email"].(string)
	if !ok {
		return "", "", "", errors.New("email no válido")
	}

	role, ok = claims["role"].(string)
	if !ok {
		return "", "", "", errors.New("role no válido")
	}

	return userID, email, role, nil
}
