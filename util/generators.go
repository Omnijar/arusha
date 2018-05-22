package util

import (
	"github.com/google/uuid"
)

const (
	// TokenLength of random tokens used by the service.
	TokenLength = 64
)

// GenerateRandomUUID generates a random UUID in string format.
func GenerateRandomUUID() string {
	return uuid.New().String()
}

// GenerateRandomToken generates a random alphanumeric token.
func GenerateRandomToken() string {
	return RandomAlphaNumeric(TokenLength)
}
