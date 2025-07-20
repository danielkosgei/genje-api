package middleware

import (
	"net/http"
	"strings"
)

// APIVersion middleware handles API versioning
func APIVersion() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Support version in header: Accept: application/vnd.genje.v1+json
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "application/vnd.genje.v") {
				// Extract version from Accept header
				// This allows for content negotiation based versioning
				w.Header().Set("Content-Type", "application/vnd.genje.v1+json")
			} else {
				w.Header().Set("Content-Type", "application/json")
			}

			// Add API version to response headers
			w.Header().Set("X-API-Version", "v1")

			next.ServeHTTP(w, r)
		})
	}
}
