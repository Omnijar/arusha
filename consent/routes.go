package consent

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/omnijar/arusha/util"
)

const (
	// ConsentPath for recording user's consents.
	ConsentPath = "/consent"
	// ConsentChallengeParameter in URL query.
	ConsentChallengeParameter = "consent_challenge"
)

// RouteHandler manages the handling of routes for consent data.
type RouteHandler struct{}

// NewRouteHandler creates a new auth route handler.
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{}
}

// SetRoutes sets the routes for authentication endpoints.
func (h *RouteHandler) SetRoutes(r *httprouter.Router) {
	r.GET(ConsentPath, h.Get)
}

// Get the user's consent.
//
// NOTE: Currently, we don't care about consent, because we don't care about third parties
// at the moment. In short, consent will always succeed, because we assume that it'll always
// happen for the root client.
func (h *RouteHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	challenge := r.URL.Query().Get(ConsentChallengeParameter)
	if challenge == "" {
		util.RespondMissingQueryParameterError(w, ConsentChallengeParameter)
		return
	}

	log.Printf("Blindly accepting consent request %s", challenge)
	response, err := util.BlindlyAcceptConsentRequest(challenge)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, response.RedirectURL, http.StatusFound)
}
