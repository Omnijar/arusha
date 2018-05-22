package users

import (
	"errors"
	"strings"

	"gitlab.com/omnijar/arusha/util"
)

// UserResource identifies an user. It contains the ID, name(s) and email(s) of an user.
// FIXME: Email and Verified fields should merge into an array.
type UserResource struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Verified  bool   `json:"verified"`
	Firstname string `json:"firstName"`
	Lastname  string `json:"lastName,omitempty"`
}

// Validate validates the user account for possible errors.
func (u *UserResource) Validate() error {
	u.ID = strings.ToUpper(u.ID)
	u.Email = strings.ToLower(u.Email)
	u.Verified = false // User shouldn't be able to do this.

	if !util.IsValidEmail(u.Email) {
		return errors.New("account: invalid email")
	}

	if u.Firstname == "" {
		return errors.New("account: first name cannot be empty for user account")
	}

	return nil
}
