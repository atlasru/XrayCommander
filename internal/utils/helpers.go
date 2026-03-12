package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new VLESS UUID
func GenerateUUID() string {
	return uuid.Must(uuid.NewRandom()).String()
}

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// MaskUUID masks a UUID for display (security)
func MaskUUID(id string) string {
	if len(id) < 8 {
		return "****"
	}
	return id[:4] + "****" + id[len(id)-4:]
}
