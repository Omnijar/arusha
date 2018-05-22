package accesscontrol

import (
	"errors"
	"strings"
)

// Scope represents a single scope. Scope names are their unique identifiers.
type Scope struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	URI         string `json:"uri"`
	Description string `json:"description"`
}

// ValidateMethodAndURI of this scope.
func (s *Scope) ValidateMethodAndURI() error {
	s.Method = strings.ToUpper(strings.TrimSpace(s.Method))
	if bitFieldForMethod(s.Method) == 0 {
		return errors.New("scope: invalid HTTP method")
	}

	s.URI = strings.TrimSpace(s.URI)

	if !strings.HasPrefix(s.URI, "/") || strings.TrimSpace(strings.Trim(s.URI, "/")) == "" {
		return errors.New("scope: invalid URI for scope")
	}

	return nil
}

// Validate the scope for possible errors.
func (s *Scope) Validate() error {
	if len(s.Name) < 4 {
		return errors.New("scope: invalid name for scope")
	}

	s.Name = strings.ToLower(s.Name)
	return s.ValidateMethodAndURI()
}
