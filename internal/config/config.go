package config

import (
	"os"
)

// Config contiene la configuración de la aplicación
type Config struct {
	Port            string
	MongoURI        string
	DBName          string
	GinMode         string
	JWTSecret       string
	AppURL          string // URL base pública del servidor (para imágenes locales)
	ShutdownTimeout int    // segundos para graceful shutdown

	// Mercado Pago
	MPAccessToken   string // Token de acceso (vacío = MP desactivado)
	MPWebhookSecret string // Clave para validar firma HMAC de webhooks
	MPSuccessURL    string // URL de redirección tras pago exitoso
	MPFailureURL    string // URL de redirección tras pago fallido
	MPPendingURL    string // URL de redirección cuando el pago queda pendiente
	MPCurrencyID    string // Código de moneda (ARS, MXN, BRL, etc.)
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8081"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:          getEnv("DB_NAME", "ecommerce"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		AppURL:          getEnv("APP_URL", "http://localhost:8080"),
		ShutdownTimeout: getEnvInt("SHUTDOWN_TIMEOUT_SECS", 10),

		MPAccessToken:   getEnv("MP_ACCESS_TOKEN", ""),
		MPWebhookSecret: getEnv("MP_WEBHOOK_SECRET", ""),
		MPSuccessURL:    getEnv("MP_SUCCESS_URL", "http://localhost:3000/checkout/success"),
		MPFailureURL:    getEnv("MP_FAILURE_URL", "http://localhost:3000/checkout/failure"),
		MPPendingURL:    getEnv("MP_PENDING_URL", "http://localhost:3000/checkout/pending"),
		MPCurrencyID:    getEnv("MP_CURRENCY_ID", "ARS"),
	}
}

func getEnvInt(key string, defaultValue int) int {
	v := getEnv(key, "")
	if v == "" {
		return defaultValue
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return defaultValue
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
