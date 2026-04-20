package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashPassword genera un hash SHA256 de una contraseña
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword verifica si una contraseña coincide con su hash
func VerifyPassword(password string, hash string) bool {
	return HashPassword(password) == hash
}
