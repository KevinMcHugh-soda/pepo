package middleware

import (
	"context"
	"net/http"
)

// RequestContextKey is the key used to store the HTTP request in context
type RequestContextKey string

const HTTPRequestKey RequestContextKey = "http_request"

// AddRequestToContext middleware adds the HTTP request to the context
// This allows handlers to access the original request for content negotiation
func AddRequestToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the request to the context
		ctx := context.WithValue(r.Context(), HTTPRequestKey, r)

		// Create a new request with the updated context
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
