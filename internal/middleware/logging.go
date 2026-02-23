package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		reqID, _ := r.Context().Value(RequestIDKey).(string)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", sw.status).
			Int("size", sw.size).
			Dur("duration", time.Since(start)).
			Str("request_id", reqID).
			Str("remote", r.RemoteAddr).
			Msg("request")
	})
}
