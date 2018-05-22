package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gitlab.com/omnijar/arusha/util"
)

const (
	// AuthPath is the GET path for auth URL and POST path for registering user credentials.
	AuthPath = "/auth"
	// TokenPath is the POST path for getting token from an authorization code.
	TokenPath = AuthPath + "/token"
	// VerifyEmailPath for verifying user's email.
	VerifyEmailPath = AuthPath + "/verify"
	// ResetSecretPath is the POST/PUT path for resetting secrets.
	ResetSecretPath = AuthPath + "/secrets/reset"
	// SessionPath is the GET,POST and PUT path for authenticating users.
	SessionPath = AuthPath + "/session"
	// ArushaAuthTokenHeader for responding with the auth token.
	ArushaAuthTokenHeader = "X-Arusha-Auth-Token"
	// ArushaRefreshTokenHeader for refreshing expired auth tokens.
	ArushaRefreshTokenHeader = "X-Arusha-Refresh-Token"
	// LoginChallengeParameter in URL query.
	LoginChallengeParameter = "login_challenge"
)

var (
	controller = &Controller{}
)

// RouteHandler manages all of the routes for authentication services.
type RouteHandler struct{}

// NewRouteHandler creates a new auth route handler.
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{}
}

// SetRoutes sets the routes for authentication endpoints.
func (h *RouteHandler) SetRoutes(r *httprouter.Router) {
	r.OPTIONS(AuthPath, util.PassEmptyBody)
	r.GET(AuthPath, h.GetAuthURL)
	r.POST(AuthPath, h.Add)
	r.OPTIONS(TokenPath, util.PassEmptyBody)
	r.GET(TokenPath, h.GetSessionToken)
	r.OPTIONS(VerifyEmailPath, util.PassEmptyBody)
	r.POST(VerifyEmailPath, h.VerifyEmailToken)
	r.OPTIONS(ResetSecretPath, util.PassEmptyBody)
	r.POST(ResetSecretPath, h.VerifySecretReset)
	r.PUT(ResetSecretPath, h.InitiateSecretReset)
	r.OPTIONS(SessionPath, util.PassEmptyBody)
	r.GET(SessionPath, h.GetSession)
	r.POST(SessionPath, h.Login)
	r.DELETE(SessionPath, h.Logout)
}

// GetAuthURL for the service.
func (h *RouteHandler) GetAuthURL(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	authURL, err := util.GetAuthURL()
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusInternalServerError)
	}

	w.Write([]byte(fmt.Sprintf(`{"status": "ok", "url": "%s"}`, *authURL)))
}

// Add a credential to the service.
func (h *RouteHandler) Add(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var credential Credential
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	resource, err := controller.Add(credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}

// VerifyEmailToken for verifying user's emails.
func (h *RouteHandler) VerifyEmailToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var credential Credential
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	resource, err := controller.VerifyEmailToken(credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resource)
}

// VerifySecretReset verifies the token and sets the secret to the provided user secret.
func (h *RouteHandler) VerifySecretReset(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var credential Credential
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	if err := controller.ResetSecret(credential); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	util.RespondHTTPStatusOK(w)
}

// InitiateSecretReset initiates the secret reset process associated with a credential.
func (h *RouteHandler) InitiateSecretReset(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var credential Credential
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	if err := controller.InitiateSecretReset(credential); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	util.RespondHTTPStatusOK(w)
}

// Login user to session.
func (h *RouteHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	challenge := r.URL.Query().Get(LoginChallengeParameter)
	if challenge == "" {
		util.RespondMissingQueryParameterError(w, LoginChallengeParameter)
		return
	}

	var credential Credential
	if r.Body == nil {
		util.RespondHTTPError(w, util.ErrorHTTPNoBody, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	var response *util.HydraRedirectResponse
	resource, err := controller.Login(credential)

	if err != nil {
		log.Printf("Rejecting login request %s (%s)", challenge, err.Error())
		response, err = util.RejectLoginRequest(challenge, err.Error())
	} else {
		log.Printf("Accepting login request %s from user %s", challenge, resource.ID)
		response, err = util.AcceptLoginRequest(challenge, resource.ID)
	}

	if err != nil {
		util.RespondHTTPError(w, err, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

// GetSession is the browser-directed page. Here, we check if the login ID already has an active token,
// which we can reuse. If a session exists, then we issue a redirect, otherwise we stay put.
func (h *RouteHandler) GetSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	challenge := r.URL.Query().Get(LoginChallengeParameter)
	if challenge == "" {
		util.RespondMissingQueryParameterError(w, LoginChallengeParameter)
		return
	}

	response, err := util.GetLoginRequest(challenge)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusInternalServerError)
		return
	}

	if response != nil {
		http.Redirect(w, r, response.RedirectURL, http.StatusFound)
		return
	}
}

// GetSessionToken for the given authorization code in URL query.
func (h *RouteHandler) GetSessionToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code := r.URL.Query().Get("code")
	if code == "" {
		util.RespondMissingQueryParameterError(w, "code")
		return
	}

	token, err := util.GetToken(code)
	if err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(token)
}

// Logout a user from a session.
func (h *RouteHandler) Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := r.Header.Get(ArushaAuthTokenHeader)
	if token == "" {
		util.RespondHTTPError(w, util.ErrorMissingToken, http.StatusBadRequest)
		return
	}

	if err := util.RevokeToken(token); err != nil {
		util.RespondHTTPError(w, err, http.StatusBadRequest)
		return
	}

	util.RespondHTTPStatusOK(w)
}
