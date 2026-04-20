package services

import "errors"

// Errores de negocio
var (
	ErrInvalidEmail       = errors.New("email inválido")
	ErrInvalidPassword    = errors.New("contraseña inválida")
	ErrEmailAlreadyExists = errors.New("el email ya existe")
	ErrUserNotFound       = errors.New("usuario no encontrado")
	ErrProductNotFound    = errors.New("producto no encontrado")
	ErrInsufficientStock  = errors.New("stock insuficiente")
	ErrOrderNotFound      = errors.New("orden no encontrada")
	ErrInvalidOrderStatus = errors.New("estado de orden inválido")

	// Errores de pagos
	ErrPaymentNotConfigured = errors.New("servicio de pagos no configurado")
	ErrPaymentNotApproved   = errors.New("pago no aprobado")
)
