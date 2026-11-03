package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type NomineeHandler struct {
	repo *repository.NomineeRepo
}

func NewNomineeHandler(repo *repository.NomineeRepo) *NomineeHandler {
	return &NomineeHandler{repo: repo}
}

// ListNominees serves GET /v1/nominees with optional party, county, heading and q filters.
func (h *NomineeHandler) ListNominees(w http.ResponseWriter, r *http.Request) {
	filter, err := parseNomineeFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	nominees, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("list nominees failed")
		writeError(w, http.StatusInternalServerError, "failed to list nominees")
		return
	}
	if nominees == nil {
		nominees = []models.Nominee{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(nominees, total, filter.Limit, filter.Offset))
}

// ListNomineeHeadings serves GET /v1/nominees/headings with each heading and its nominee count.
func (h *NomineeHandler) ListNomineeHeadings(w http.ResponseWriter, r *http.Request) {
	headings, err := h.repo.ListHeadings(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("list nominee headings failed")
		writeError(w, http.StatusInternalServerError, "failed to list nominee headings")
		return
	}
	if headings == nil {
		headings = []models.NomineeHeading{}
	}
	writeJSON(w, http.StatusOK, headings)
}
