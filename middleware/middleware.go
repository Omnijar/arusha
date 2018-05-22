package middleware

import (
	"net/http"
)

// Middleware provides functions for middleware.
type Middleware struct {
	next http.Handler
}

// New makes a constructor for our middleware type since its fields are not exported (in lowercase)
func New(next http.Handler) *Middleware {
	return &Middleware{next: next}
}
