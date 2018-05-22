package util

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	// ErrorHTTPNoBody represents an error in HTTP request without a body.
	ErrorHTTPNoBody = errors.New("missing request body")
	// ErrorInvalidToken for requests with invalid auth tokens.
	ErrorInvalidToken = errors.New("invalid token in header")
	// ErrorOAuthNotInitialized for uninitialized Hydra.
	ErrorOAuthNotInitialized = errors.New("auth engine not initialized")
	// ErrorOAuthFetch occurs when any hydra request fails.
	ErrorOAuthFetch = errors.New("error getting auth information")
	// ErrorMissingToken represents missing auth token in header.
	ErrorMissingToken = errors.New("missing access token for resource")
	// ErrorRBACNotInitialized for uninitialized Keto.
	ErrorRBACNotInitialized = errors.New("access control engine not initialized")
)

// RespondMissingQueryParameterError for a request.
func RespondMissingQueryParameterError(w http.ResponseWriter, param string) {
	RespondHTTPError(w, fmt.Errorf("missing query parameter '%s' in URL", param), http.StatusBadRequest)
}

// RespondHTTPStatusOK responds with an OK status in HTTP JSON format.
func RespondHTTPStatusOK(w http.ResponseWriter) {
	w.Write([]byte(`{"status": "ok"}`))
}

// RespondHTTPError responds with an error message in HTTP JSON format with the given status code.
func RespondHTTPError(w http.ResponseWriter, err error, code int) {
	http.Error(w, fmt.Sprintf(`{"status": "error", "message": "%s"}`, err.Error()), code)
}

// PassEmptyBody for the specified route.
func PassEmptyBody(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}
