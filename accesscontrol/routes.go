package accesscontrol

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/omnijar/arusha/util"
)

const (
	// RolesPath for creating/modifying roles.
	RolesPath = "/roles"
	// RolePath for an individual role.
	RolePath = RolesPath + "/:id"
	// ScopesPath for creating/modifying scopes.
	ScopesPath = "/scopes"
	// ScopesSelfPath for initializing scopes of this Arusha instance for the first time.
	ScopesSelfPath = ScopesPath + "/init"
	// ScopesAuthorizePath for authorizing a request to the given URL.
	ScopesAuthorizePath = ScopesPath + "/authorize"
)

var (
	controller = &Controller{}
)

// RouteHandler manages all of the routes for access control.
type RouteHandler struct{}

// NewRouteHandler creates a new access control route handler.
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{}
}

// SetRoutes sets the routes for access control.
func (h *RouteHandler) SetRoutes(r *httprouter.Router) {
	r.OPTIONS(ScopesPath, util.PassEmptyBody)
	r.GET(ScopesPath, h.GetScopes)
	r.OPTIONS(ScopesSelfPath, util.PassEmptyBody)
	r.POST(ScopesSelfPath, h.InitializeScopes)
	r.OPTIONS(ScopesAuthorizePath, util.PassEmptyBody)
	r.POST(ScopesAuthorizePath, h.AuthorizeAction)
	r.OPTIONS(RolesPath, util.PassEmptyBody)
	r.POST(RolesPath, h.CreateRole)
	r.GET(RolesPath, h.ListRoles)
	r.OPTIONS(RolePath, util.PassEmptyBody)
	r.GET(RolePath, h.GetRole)
	r.PUT(RolePath, h.UpdateRole)
	r.DELETE(RolePath, h.DeleteRole)
}

// InitializeScopes for this Arusha instance. This can only be done once.
func (h *RouteHandler) InitializeScopes(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var scopes []Scope
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&scopes)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	token, err := controller.InitializeScopes(scopes)
	if err != nil {
		log.Println("error initializing scopes:", err)
		controller.Reset()
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"status": "ok", "token": "%s"}`, *token)))
}

// GetScopes from this instance.
func (h *RouteHandler) GetScopes(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	scopes, err := controller.GetScopes()
	if err != nil {
		log.Println("error getting scopes:", err)
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(scopes)
}

// AuthorizeAction made by the subject. This checks whether the subject resolved from the
// token is allowed to carry out an action (i.e., HTTP method on a route)
func (h *RouteHandler) AuthorizeAction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	authToken := r.Header.Get("Authorization")
	if len(authToken) > 7 {
		authToken = authToken[7:]
	} else {
		authToken = ""
	}

	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	var scope Scope
	err := json.NewDecoder(r.Body).Decode(&scope)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	if err = controller.AuthorizeToken(authToken, scope); err != nil {
		log.Printf("error authorizing token for action %s %s: %s", scope.Method, scope.URI, err)
		util.RespondHTTPError(w, err, http.StatusForbidden)
		return
	}

	util.RespondHTTPStatusOK(w)
}

// GetRolesForSubject associated with the token.
func (h *RouteHandler) GetRolesForSubject(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	authToken := r.Header.Get("Authorization")
	if len(authToken) > 7 {
		authToken = authToken[7:]
	} else {
		authToken = ""
	}

	if controller.IsRootToken(authToken) {
		json.NewEncoder(w).Encode([]string{util.AdminRole})
		return
	}

	subject, err := util.AuthorizeToken(authToken)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusForbidden)
		return
	}

	roles, err := util.ListRolesForSubject(*subject)
	if err != nil {
		log.Printf("error fetching roles for subject %s: %s", *subject, err.Error())
	}

	json.NewEncoder(w).Encode(roles)
}

// CreateRole with the specified scopes and members.
func (h *RouteHandler) CreateRole(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var role Role
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	newRole, err := controller.CreateRole(role)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(newRole)
}

// UpdateRole corresponding to an ID with the given scopes and members.
func (h *RouteHandler) UpdateRole(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	roleID := params.ByName("id")
	var role Role
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	newRole, err := controller.UpdateRole(roleID, role)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(newRole)
}

// DeleteRole corresponding to an ID.
func (h *RouteHandler) DeleteRole(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	roleID := params.ByName("id")
	if err := controller.DeleteRole(roleID); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	util.RespondHTTPStatusOK(w)
}

// ListRoles registered in this instance.
func (h *RouteHandler) ListRoles(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	roles, err := controller.ListRoles()
	if err != nil {
		log.Println("error getting roles:", err)
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// GetRole corresponding to an ID.
func (h *RouteHandler) GetRole(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	roleID := params.ByName("id")
	if roleID == "self" {
		h.GetRolesForSubject(w, r, params)
		return
	}

	role, err := controller.GetRole(roleID)
	if err != nil {
		log.Println("error getting role:", err)
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(role)
}
