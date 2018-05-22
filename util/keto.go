package util

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ory/keto/sdk/go/keto"
	ketoAPI "github.com/ory/keto/sdk/go/keto/swagger"
)

const (
	// AdminRole with exclusive privileges to all scopes.
	AdminRole = "admin"
	// AdminDescription for admin role.
	AdminDescription = "Special policy for admin to do anything on all scopes"
	// StubAction for keto policies. We don't use keto's "actions". Since Arusha's scopes
	// themselves are unique, we differentiate the actions with scope names. Hence, we use
	// a default action for all keto calls.
	StubAction = "perform"
	// RolePolicyPrefix for policy IDs associated with roles.
	RolePolicyPrefix = "arusha.role."
	// EnvKetoClusterURL for Keto's private/public URL.
	EnvKetoClusterURL = "KETO_CLUSTER_URL"
)

var (
	ketoEndpoint string
	ketoClient   *keto.CodeGenSDK
)

// InitializeKetoClient for communicating with keto.
func InitializeKetoClient() error {
	endpoint := os.Getenv(EnvKetoClusterURL)
	_, err := url.Parse(endpoint)
	if err != nil || endpoint == "" {
		return errors.New(EnvKetoClusterURL + " variable not configured or invalid")
	}

	ketoEndpoint = strings.TrimRight(endpoint, "/")

	ketoClient, err = keto.NewCodeGenSDK(&keto.Configuration{
		EndpointURL: ketoEndpoint,
	})

	if err != nil {
		return errors.New("keto: error initializing keto client: " + err.Error())
	}

	return nil
}

// CreateRole creates a role/policy pair in Keto.
//
// Some terminology:
// - Role(A) - Arusha's role which contains a role's unique name, its scopes, member IDs and description.
// - Role(K) - Keto's role which has the role ID and list of members (or subjects).
// - Policy(K) - Keto's policy which contains a description, list of scopes (or resources) and subjects
// for those scopes.
//
// Role(K) and Policy(K) could have any number of subjects, and Role(K) could be a subject in Policy(K).
// To keep things simple, we restrict a Policy(K) to a single subject and that subject is Role(K).
// Policy(K) will have the corresponding scopes required by Role(A), and whenever we update Role(A),
// this function will update Role(K) and Policy(K) correspondingly.
func CreateRole(id, description string, members, scopes []string) error {
	if oauth2Config == nil {
		return ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return ErrorRBACNotInitialized
	}

	_, response, err := ketoClient.RoleApi.CreateRole(ketoAPI.Role{
		Id:      id,
		Members: members,
	})

	if err != nil || response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("keto: error creating role '%s': %s", id, err)
	}

	policy, response, err := ketoClient.PolicyApi.CreatePolicy(ketoAPI.Policy{
		Id:          RolePolicyPrefix + id,
		Actions:     []string{StubAction},
		Description: description,
		Resources:   scopes,
		Effect:      "allow",
	})

	if err != nil || response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("keto: error creating policy for role '%s': %s", id, err)
	}

	log.Printf("keto: created policy %s for role %s", policy.Id, id)

	return nil
}

// UpdateRole for a given ID with a description, list of members and scopes.
func UpdateRole(roleID, id, description string, members, scopes []string) error {
	if oauth2Config == nil {
		return ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return ErrorRBACNotInitialized
	}

	if roleID == AdminRole { // admin role's name and scopes cannot be updated.
		id = AdminRole
		description = AdminDescription
		scopes = oauth2Config.Scopes
	}

	if err := DeleteRole(roleID); err != nil {
		return err
	}

	if err := CreateRole(id, description, members, scopes); err != nil {
		return err
	}

	return nil
}

// DeleteRole corresponding to an ID.
func DeleteRole(id string) error {
	if oauth2Config == nil {
		return ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return ErrorRBACNotInitialized
	}

	response, err := ketoClient.RoleApi.DeleteRole(id)
	if err != nil || response.StatusCode >= http.StatusBadRequest {
		log.Printf("keto: error deleting role on update: %s", err)
		return errors.New("keto: role doesn't exist")
	}

	response, err = ketoClient.PolicyApi.DeletePolicy(RolePolicyPrefix + id)
	if err != nil || response.StatusCode >= http.StatusBadRequest {
		log.Printf("keto: error deleting policy for role '%s': %s", id, err)
		return errors.New("keto: policy doesn't exist")
	}

	log.Printf("keto: deleted policy for role %s", id)
	return nil
}

// ListRolesAndPolicies from keto for constructing Arusha roles.
func ListRolesAndPolicies() ([]ketoAPI.Role, []ketoAPI.Policy, error) {
	if oauth2Config == nil {
		return nil, nil, ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return nil, nil, ErrorRBACNotInitialized
	}

	policies, response, err := ketoClient.PolicyApi.ListPolicies(0, 500) // FIXME: Won't be scalable
	if err != nil || response.StatusCode >= http.StatusBadRequest {
		return nil, nil, fmt.Errorf("keto: error fetching role policies: %s", err)
	}

	rolePolicies := *new([]ketoAPI.Policy)
	roles := *new([]ketoAPI.Role)
	for _, policy := range policies {
		if !strings.HasPrefix(policy.Id, RolePolicyPrefix) {
			continue
		}

		roleID := policy.Id[len(RolePolicyPrefix):]
		role, response, err := ketoClient.RoleApi.GetRole(roleID)
		if err != nil || response.StatusCode >= http.StatusBadRequest {
			return nil, nil, fmt.Errorf("error fetching role %s: %s", roleID, err)
		}

		roles = append(roles, *role)
		rolePolicies = append(rolePolicies, policy)
	}

	return roles, policies, nil
}

// GetRolePolicyPair for constructing an Arusha role.
func GetRolePolicyPair(roleID string) (*ketoAPI.Role, *ketoAPI.Policy, error) {
	if oauth2Config == nil {
		return nil, nil, ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return nil, nil, ErrorRBACNotInitialized
	}

	role, response, err := ketoClient.RoleApi.GetRole(roleID)
	if err != nil || response.StatusCode >= http.StatusBadRequest {
		return nil, nil, fmt.Errorf("error fetching role %s: %s", roleID, err)
	}

	policy, response, err := ketoClient.PolicyApi.GetPolicy(RolePolicyPrefix + roleID)
	if err != nil || response.StatusCode >= http.StatusBadRequest {
		return nil, nil, fmt.Errorf("keto: error fetching policy for role %s: %s", roleID, err)
	}

	return role, policy, nil
}

// CreateAdminRole with the scopes initialized in hydra.
func CreateAdminRole() error {
	if oauth2Config == nil {
		return ErrorOAuthNotInitialized
	}

	if ketoClient == nil {
		return ErrorRBACNotInitialized
	}

	// Delete existing admin role and policy before creating one.
	_, _ = ketoClient.RoleApi.DeleteRole(AdminRole)
	_, err := ketoClient.PolicyApi.DeletePolicy(RolePolicyPrefix + AdminRole)

	if err = CreateRole(AdminRole, AdminDescription, []string{}, oauth2Config.Scopes); err != nil {
		return err
	}

	return nil
}

// IsSubjectAuthorized for the given scope?
func IsSubjectAuthorized(subject, scope string) (bool, error) {
	if ketoClient == nil {
		return false, ErrorRBACNotInitialized
	}

	log.Println("keto: authorizing subject %s for performing action on scope %s", subject, scope)

	response, _, err := ketoClient.WardenApi.IsSubjectAuthorized(ketoAPI.WardenSubjectAuthorizationRequest{
		Action:   StubAction,
		Resource: scope,
		Subject:  subject,
	})

	if err != nil {
		return false, errors.New("keto: error authorizing subject: " + err.Error())
	}

	return response.Allowed, nil
}

// ListRolesForSubject returns the role IDs for a subject.
func ListRolesForSubject(subject string) ([]string, error) {
	if ketoClient == nil {
		return []string{}, ErrorRBACNotInitialized
	}

	roles, _, err := ketoClient.RoleApi.ListRoles(subject, 500, 0) // FIXME: Won't be scalable
	if err != nil {
		return []string{}, errors.New("keto: error fetching roles: " + err.Error())
	}

	roleIds := *new([]string)
	for _, role := range roles {
		roleIds = append(roleIds, role.Id)
	}

	return roleIds, nil
}
