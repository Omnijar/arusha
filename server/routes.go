package server

import (
	"github.com/julienschmidt/httprouter"
	"gitlab.com/omnijar/arusha/accesscontrol"
	"gitlab.com/omnijar/arusha/auth"
	"gitlab.com/omnijar/arusha/consent"
	"gitlab.com/omnijar/arusha/users"
)

// RouterHandler describes how router handlers must look.
type RouterHandler interface {
	SetRoutes(r *httprouter.Router)
}

// RouteHandler contains the domain-based route handlers for the HTTP service.
type RouteHandler struct {
	Access  *accesscontrol.RouteHandler
	Auth    *auth.RouteHandler
	Consent *consent.RouteHandler
	Users   *users.RouteHandler
}

func (h *RouteHandler) registerRoutes(router *httprouter.Router) {
	h.Access = accesscontrol.NewRouteHandler()
	h.Auth = auth.NewRouteHandler()
	h.Consent = consent.NewRouteHandler()
	h.Users = users.NewRouteHandler()

	h.Access.SetRoutes(router)
	h.Auth.SetRoutes(router)
	h.Consent.SetRoutes(router)
	h.Users.SetRoutes(router)
}
