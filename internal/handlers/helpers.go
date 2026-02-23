package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"jalada/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}
