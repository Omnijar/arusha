package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/omnijar/arusha/util"
)

const (
	// MinSecretLength is the minimum length required for user's secret.
	MinSecretLength = 8
)

// Credential has a number of purposes in an incoming payload.
//
// - For verifying an user's email, the Token field is set.
// - While setting the password for the first time, this has either the Email/ID, along
// with the user's Secret. Post-auth, the hash of the Secret is stored in the store.
// - During authentication, the stored hash is compared with the hash of the user's secret
// in the incoming payload.
// - When requesting a password reset, only the Email field is set (and an email is sent
// with the verification link).
// - When resetting the password, the Token (obtained from verification link) and (new) Secret
// are set. If the token is valid and hasn't expired, then the secret is updated.
type Credential struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Secret string `json:"secret"`
	Token  string `json:"token"`
}

// ValidateEmail validates the email field.
func (c *Credential) ValidateEmail() error {
	c.Email = strings.ToLower(c.Email)
	if !util.IsValidEmail(c.Email) {
		return errors.New("credential: invalid email")
	}

	return nil
}

// ValidateSecret validates the secret field.
func (c *Credential) ValidateSecret() error {
	if len(c.Secret) < MinSecretLength {
		return errors.New("credential: secret is too short")
	}

	return nil
}

// ValidateToken validates the token field.
func (c *Credential) ValidateToken() error {
	if len(c.Token) < util.TokenLength {
		return errors.New("auth: invalid token")
	}

	return nil
}

// Validate validates the credential for possible errors.
func (c *Credential) Validate() error {
	if c.ID != "" {
		c.ID = strings.ToUpper(c.ID)
	} else if c.Email != "" {
		if err := c.ValidateEmail(); err != nil {
			return err
		}
	} else {
		return errors.New("credential: user ID or email required")
	}

	return c.ValidateSecret()
}

// HashSecret hashes a given string as a secret with SHA-256.
func HashSecret(secret string) string {
	bytes := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", bytes)
}

// CheckSecretHash compares a secret and a hash to ensure they match.
func CheckSecretHash(secret, hash string) bool {
	bytes := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", bytes) == hash
}
