package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/omnijar/arusha/util"
)

// Our middleware handler
func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We can modify the request here; for simplicity, we will just log a message
	log.Printf("Method: %s, URI: %s\n", r.Method, r.RequestURI)

	includeVersion(&w)
	specifyContent(&w)
	enableCors(&w)

	m.next.ServeHTTP(w, r)
}

func specifyContent(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
}

func includeVersion(w *http.ResponseWriter) {
	(*w).Header().Set("Server", "Arusha/0.2.0 (Unix)")
}

// EnableCors ensures Cores is permitted on HTTP requests.
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

type key int

const requestIDKey key = 0

func newContextWithRequestID(ctx context.Context, req *http.Request) context.Context {
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = util.GenerateRandomUUID()
	}

	return context.WithValue(ctx, requestIDKey, reqID)
}

func requestIDFromContext(ctx context.Context) string {
	return ctx.Value(requestIDKey).(string)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := newContextWithRequestID(req.Context(), req)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func handler(rw http.ResponseWriter, req *http.Request) {
	reqID := requestIDFromContext(req.Context())
	fmt.Fprintf(rw, "Hello request ID %v\n", reqID)
}
