package users

import (
	"errors"

	"gitlab.com/omnijar/arusha/util"
)

const (
	usersPath  = "/users"
	emailsPath = "/emails"

	// VaultResetTokenPath has all the active tokens for resetting secret.
	VaultResetTokenPath = "/reset-tokens"
	// VaultEmailVerifyPath has all the active tokens for verifying email.
	VaultEmailVerifyPath = "/email-tokens"
)

var (
	users = make(map[string]UserResource)
)

// Controller is a controller for managing user functions.
type Controller struct{}

// NewController for managing user events.
func NewController() *Controller {
	return &Controller{}
}

// Add a user resource to the system.
func (c *Controller) Add(user UserResource) (*UserResource, error) {
	// Generate random ID for the user.
	user.ID = util.GenerateRandomUUID()
	// Validate their data.
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Check whether a resource already exists for the email.
	emailVault := util.GetVaultClient(emailsPath)

	var userID string
	if userExists := emailVault.Get(user.Email, &userID); userExists {
		return nil, errors.New("users: email already exists. resource cannot be addded")
	}

	// Generate a random token and send verification mail.
	token := util.GenerateRandomToken()
	tokenVault := util.GetVaultClient(VaultEmailVerifyPath)
	tokenVault.Set(token, user.Email)
	go util.SendVerificationMail(user.Email, token)

	// Create user data.
	usersVault := util.GetVaultClient(usersPath)
	users[user.ID] = user // FIXME: Remove this!
	usersVault.Set(user.ID, user)
	emailVault.Set(user.Email, user.ID)

	return &user, nil
}

// Update an user resource within the system.
func (c *Controller) Update(newResource UserResource) (*UserResource, error) {
	// Validate new data.
	if err := newResource.Validate(); err != nil {
		return nil, err
	}

	// Ensure that data already exists for resource.
	usersVault := util.GetVaultClient(usersPath)

	var oldResource UserResource
	if resourceExists := usersVault.Get(newResource.ID, &oldResource); !resourceExists {
		return nil, errors.New("users: resource doesn't exist. resource cannot be updated")
	}

	// If the email is new, then send verification mail and update the store.
	if oldResource.Email != newResource.Email {
		token := util.GenerateRandomToken()
		tokenVault := util.GetVaultClient(VaultEmailVerifyPath)
		tokenVault.Set(token, newResource.Email)
		go util.SendVerificationMail(newResource.Email, token)

		emailVault := util.GetVaultClient(emailsPath)
		emailVault.Remove(oldResource.Email)
		emailVault.Set(newResource.Email, newResource.ID)
	} else {
		newResource.Verified = oldResource.Verified
	}

	users[newResource.ID] = newResource // FIXME: Remove this!
	usersVault.Set(newResource.ID, newResource)
	return &newResource, nil
}

// FindUserResourceByID gets an user resource from the system based on the ID.
func (c *Controller) FindUserResourceByID(id string) (*UserResource, error) {
	vault := util.GetVaultClient(usersPath)

	var resource UserResource
	if resourceExists := vault.Get(id, &resource); resourceExists {
		users[id] = resource // FIXME: Remove this
		return &resource, nil
	}

	return nil, errors.New("users: resource doesn't exist for ID")
}

// FindUserResourceByEmail gets an user resource from the system based on their email.
func (c *Controller) FindUserResourceByEmail(email string) (*UserResource, error) {
	vault := util.GetVaultClient(emailsPath)

	var userID string
	if emailExists := vault.Get(email, &userID); emailExists {
		return c.FindUserResourceByID(userID)
	}

	return nil, errors.New("users: resource doesn't exist for email")
}

// FetchAllResources from this instance.
// FIXME: How do we paginate? Vault doesn't support pagination. So, we should probably
// go for a database?
func (c *Controller) FetchAllResources() (map[string]UserResource, error) {
	return users, nil
}

// RemoveUserResource corresponding to the given ID.
func (c *Controller) RemoveUserResource(id string) (*UserResource, error) {
	usersVault := util.GetVaultClient(usersPath)
	emailVault := util.GetVaultClient(emailsPath)

	resource, err := c.FindUserResourceByID(id)
	if err != nil {
		return nil, err
	}

	delete(users, resource.ID)
	usersVault.Remove(resource.ID)
	emailVault.Remove(resource.Email)

	return resource, nil
}

// VerifyEmail marks the given email as verified (if it exists).
func (c *Controller) VerifyEmail(email string) {
	user, err := c.FindUserResourceByEmail(email)
	if err != nil {
		return
	}

	user.Verified = true

	vault := util.GetVaultClient(usersPath)
	vault.Set(user.ID, user)
}
