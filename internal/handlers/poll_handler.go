package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/rs/zerolog/log"

	"jalada/internal/models"
	"jalada/internal/repository"
)

const (
	pollDefaultLimit = 20
	pollMaxLimit     = 100
)

// PollHandler serves the published poll list.
type PollHandler struct {
	repo *repository.PollRepo
}

func NewPollHandler(repo *repository.PollRepo) *PollHandler {
	return &PollHandler{repo: repo}
}

// ListPolls serves GET /v1/polls. Optional race filter, limit (1-100) and offset.
func (h *PollHandler) ListPolls(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	race := q.Get("race")
	if race != "" && !slices.Contains(models.PollRaces, race) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("race must be one of %v", models.PollRaces))
		return
	}

	limit := pollDefaultLimit
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > pollMaxLimit {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("limit must be a whole number from 1 to %d", pollMaxLimit))
			return
		}
		limit = n
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "offset must be a whole number of 0 or more")
			return
		}
		offset = n
	}

	polls, total, err := h.repo.List(r.Context(), race, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("list polls failed")
		writeError(w, http.StatusInternalServerError, "failed to list polls")
		return
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(polls, total, limit, offset))
}
