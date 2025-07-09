package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/cors"
)

func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := generateRequestID()
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
} 