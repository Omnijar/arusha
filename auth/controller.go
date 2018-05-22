package auth

import (
	"errors"

	"gitlab.com/omnijar/arusha/users"
	"gitlab.com/omnijar/arusha/util"
)

const (
	credentialsPath = "/credentials"
)

var (
	usersController = users.NewController()
)

// Controller blah
type Controller struct{}

// FIXME: Instead of hashing the secret directly, prefer salted hashing.

// Add a credential to the auth service.
func (c *Controller) Add(credential Credential) (*users.UserResource, error) {
	if err := credential.ValidateToken(); err != nil {
		return nil, err
	}

	// FIXME: We're checking the token anyway, do we really need another route?

	emailVault := util.GetVaultClient(users.VaultEmailVerifyPath)
	var email string
	if emailExists := emailVault.Get(credential.Token, &email); !emailExists {
		return nil, errors.New("auth: invalid token")
	}

	credential.ID = ""
	credential.Email = email
	user, err := c.getResource(&credential)
	if err != nil {
		return nil, err
	}

	if !user.Verified {
		usersController.VerifyEmail(email)
		// return nil, errors.New("auth: credentials cannot be created before verifying email")
	}

	emailVault.Remove(credential.Token)

	secretVault := util.GetVaultClient(credentialsPath)

	var hash string
	if hashExists := secretVault.Get(user.ID, &hash); hashExists {
		return nil, errors.New("auth: credentials have already been created")
	}

	hash = HashSecret(credential.Secret)
	secretVault.Set(user.ID, hash)
	return user, nil
}

// VerifyEmailToken for verifying registered emails.
func (c *Controller) VerifyEmailToken(credential Credential) (*users.UserResource, error) {
	if err := credential.ValidateToken(); err != nil {
		return nil, err
	}

	vault := util.GetVaultClient(users.VaultEmailVerifyPath)

	var email string
	if emailExists := vault.Get(credential.Token, &email); !emailExists {
		return nil, errors.New("auth: invalid token")
	}

	usersController.VerifyEmail(email)

	return usersController.FindUserResourceByEmail(email)
}

// ResetSecret of a credential in the auth service.
func (c *Controller) ResetSecret(credential Credential) error {
	if err := credential.ValidateSecret(); err != nil {
		return err
	}

	if err := credential.ValidateToken(); err != nil {
		return err
	}

	tokenVault := util.GetVaultClient(users.VaultResetTokenPath)

	var userID string
	if userExists := tokenVault.Get(credential.Token, &userID); !userExists {
		return errors.New("auth: invalid token")
	}

	tokenVault.Remove(credential.Token)

	secretVault := util.GetVaultClient(credentialsPath)
	hash := HashSecret(credential.Secret)
	secretVault.Set(userID, hash)

	return nil
}

// InitiateSecretReset sends a mail to the suspect's registered email to verify their identity.
func (c *Controller) InitiateSecretReset(credential Credential) error {
	if err := credential.ValidateEmail(); err != nil {
		return err
	}

	user, err := usersController.FindUserResourceByEmail(credential.Email)
	if err != nil {
		return err
	}

	token := util.GenerateRandomToken()
	vault := util.GetVaultClient(users.VaultResetTokenPath)
	vault.Set(token, user.ID)
	go util.SendSecretResetMail(user.Email, token)

	return nil
}

// Login user to service.
func (c *Controller) Login(credential Credential) (*users.UserResource, error) {
	user, err := c.getResource(&credential)
	if err != nil {
		return nil, err
	}

	vault := util.GetVaultClient(credentialsPath)

	var hash string
	if hashExists := vault.Get(user.ID, &hash); !hashExists {
		return nil, errors.New("auth: credentials don't exist. cannot login")
	}

	if isMatchingSecret := CheckSecretHash(credential.Secret, hash); !isMatchingSecret {
		return nil, errors.New("auth: wrong secret")
	}

	return user, nil
}

// GetResource gets the user resource for the given credential.
func (c *Controller) getResource(credential *Credential) (*users.UserResource, error) {
	if err := credential.Validate(); err != nil {
		return nil, err
	}

	var user *users.UserResource
	var err error

	if credential.ID != "" {
		user, err = usersController.FindUserResourceByID(credential.ID)
	} else {
		user, err = usersController.FindUserResourceByEmail(credential.Email)
	}

	return user, err
}
