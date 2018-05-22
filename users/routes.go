package users

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/omnijar/arusha/util"
)

const (
	// UsersPath for adding new user resources.
	UsersPath = "/users"
	// UserPath for modifying a single user resource.
	UserPath = UsersPath + "/:id"
)

var (
	controller = NewController()
)

// RouteHandler manages the handling of routes for user data.
type RouteHandler struct{}

// NewRouteHandler creates a new auth route handler.
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{}
}

// SetRoutes sets the routes for authentication endpoints.
func (h *RouteHandler) SetRoutes(r *httprouter.Router) {
	r.OPTIONS(UsersPath, util.PassEmptyBody)
	r.POST(UsersPath, h.Add)
	r.GET(UsersPath, h.List)
	r.OPTIONS(UserPath, util.PassEmptyBody)
	r.GET(UserPath, h.Get)
	r.PUT(UserPath, h.Update)
	r.DELETE(UserPath, h.Remove)
}

// Get returns the existing user data.
func (h *RouteHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")

	resource, err := controller.FindUserResourceByID(id)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}

// List all user resources.
func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	resources, err := controller.FetchAllResources()
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resources)
}

// Add appends a new user resource.
func (h *RouteHandler) Add(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var user UserResource
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	resource, err := controller.Add(user)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}

// Update modifies a user resource.
func (h *RouteHandler) Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var resource UserResource
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	resource.ID = params.ByName("id")
	_, err := controller.Update(resource)

	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}

// Remove the user corresponding to the given ID.
func (h *RouteHandler) Remove(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")

	resource, err := controller.RemoveUserResource(id)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}
