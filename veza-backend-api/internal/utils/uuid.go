package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID génère un UUID v4 unique
func GenerateUUID() string {
	return uuid.New().String()
}

// IsValidUUID vérifie si une chaîne est un UUID valide
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
} 