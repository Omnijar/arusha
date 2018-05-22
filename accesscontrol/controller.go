package accesscontrol

import (
	"errors"
	"log"

	"gitlab.com/omnijar/arusha/users"
	"gitlab.com/omnijar/arusha/util"
)

var (
	allScopes    []Scope
	scopeNameMap map[string]int
	scopeTree    *ScopeRouteTree
	rootToken    *string
	// ErrorScopesInitialized occurs when scopes have already been initialized.
	ErrorScopesInitialized = errors.New("scopes have already been initialized. Please perform an update request to update them")
	usersController        = users.NewController()
)

// Controller for managing scopes for access control.
type Controller struct{}

// InitializeScopes for this controller. Each call will replace all existing scopes.
// If this method fails, then `Reset` should be called to clear unusable scopes from memory.
func (c *Controller) InitializeScopes(scopes []Scope) (*string, error) {
	if rootToken != nil {
		return nil, ErrorScopesInitialized
	}

	allScopes = *new([]Scope)
	scopeNames := *new([]string)
	scopeNameMap = make(map[string]int)
	scopeTree = NewScopeRouteTree()

	for _, scope := range scopes {
		if err := scope.Validate(); err != nil {
			return nil, err
		}

		if _, exists := scopeNameMap[scope.Name]; exists {
			return nil, errors.New("scope: " + scope.Name + " already exists")
		}

		scopeIdx := len(allScopes)
		allScopes = append(allScopes, scope)
		scopeNames = append(scopeNames, scope.Name)
		scopeNameMap[scope.Name] = scopeIdx
		scopeTree.AddRoute(scope.Method, scope.URI, scopeIdx)
	}

	if err := util.InitializeRootHydraClient(scopeNames); err != nil {
		return nil, err
	}

	if err := util.CreateAdminRole(); err != nil {
		return nil, err
	}

	token := util.RandomAlphaNumeric(64)
	rootToken = &token
	return rootToken, nil
}

// GetScopes from this controller. This requires the scopes to be initialized first.
func (c *Controller) GetScopes() ([]Scope, error) {
	if rootToken == nil {
		return nil, errors.New("scopes haven't been initialized")
	}

	return allScopes, nil
}

// Reset this controller.
func (c *Controller) Reset() {
	allScopes = nil
	scopeNameMap = nil
	scopeTree = nil
}

// AuthorizeToken for the given action.
func (c *Controller) AuthorizeToken(token string, scope Scope) error {
	if err := scope.ValidateMethodAndURI(); err != nil {
		return err
	}

	if c.IsRootToken(token) {
		return nil
	}

	scopeIndices := scopeTree.GetMatchingScopes(scope.Method, scope.URI)
	log.Printf("access: found %d scopes for %s %s", len(scopeIndices), scope.Method, scope.URI)
	if len(scopeIndices) == 0 {
		log.Printf("access: %s %s not registered for any scope. Allowing action...", scope.Method, scope.URI)
		return nil
	}

	// We have found some matching scopes, but the auth token is too short.
	if len(token) < 10 {
		return util.ErrorInvalidToken
	}

	subject, err := util.AuthorizeToken(token)
	if err != nil {
		return err
	}

	// FIXME: We're checking each scope. Is this the only way with Keto?
	for _, scopeIdx := range scopeIndices {
		scopeName := allScopes[scopeIdx].Name
		if allowed, err := util.IsSubjectAuthorized(*subject, scopeName); allowed {
			if err != nil {
				log.Printf(err.Error())
			}

			return nil
		}
	}

	return errors.New("access: invalid token or unauthorized")
}

// IsRootToken matching the given subject token?
func (c *Controller) IsRootToken(token string) bool {
	if rootToken == nil {
		log.Println("access: scopes haven't been initialized. all requests will be allowed.")
		return true
	}

	return token == *rootToken
}

// CreateRole using the given data.
func (c *Controller) CreateRole(role Role) (*Role, error) {
	if err := role.Validate(); err != nil {
		return nil, err
	}

	if err := util.CreateRole(role.ID, role.Description, role.Members, role.Scopes); err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateRole using the given data.
func (c *Controller) UpdateRole(id string, role Role) (*Role, error) {
	if err := role.Validate(); err != nil {
		return nil, err
	}

	if err := util.UpdateRole(id, role.ID, role.Description, role.Members, role.Scopes); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole using the given ID.
func (c *Controller) DeleteRole(id string) error {
	if id == util.AdminRole {
		return errors.New("admin role cannot be deleted")
	}

	if err := util.DeleteRole(id); err != nil {
		return err
	}

	return nil
}

// ListRoles from Arusha.
func (c *Controller) ListRoles() ([]Role, error) {
	roles, policies, err := util.ListRolesAndPolicies()
	if err != nil {
		return nil, err
	}

	arushaRoles := *new([]Role)
	for i := range roles {
		role := Role{
			ID:          roles[i].Id,
			Description: policies[i].Description,
			Members:     roles[i].Members,
			Scopes:      policies[i].Resources,
		}

		arushaRoles = append(arushaRoles, role)
	}

	return arushaRoles, nil
}

// GetRole corresponding to the given ID
func (c *Controller) GetRole(id string) (*Role, error) {
	role, policy, err := util.GetRolePolicyPair(id)
	if err != nil {
		return nil, err
	}

	return &Role{
		ID:          role.Id,
		Description: policy.Description,
		Members:     role.Members,
		Scopes:      policy.Resources,
	}, nil
}
