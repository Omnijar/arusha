package accesscontrol

import (
	"errors"
	"strings"
)

// Role contains the access information (scopes) for members.
type Role struct {
	ID          string   `json:"name"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	Scopes      []string `json:"scopes"`
}

// Validate this role for possible errors.
func (r *Role) Validate() error {
	r.ID = strings.ToLower(r.ID)
	if r.ID == "" {
		return errors.New("role: name should be unique and cannot be empty")
	}

	members := make(map[string]bool)
	for _, member := range r.Members {
		if _, exists := members[member]; exists {
			continue // filter duplicates
		}

		if _, err := usersController.FindUserResourceByID(member); err != nil {
			return err
		}

		members[member] = true
	}

	scopes := make(map[string]bool)
	for _, scope := range r.Scopes {
		if _, exists := scopes[scope]; exists {
			continue // filter duplicates
		}

		if _, exists := scopeNameMap[scope]; !exists {
			return errors.New("scope " + scope + " doesn't exist")
		}
	}

	return nil
}
