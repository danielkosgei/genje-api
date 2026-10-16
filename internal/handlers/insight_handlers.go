package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"jalada/internal/models"
	"jalada/internal/repository"
)

const (
	searchDefaultLimit = 5
	searchMaxLimit     = 20
	searchMaxQueryLen  = 100
	pollSummaryWindow  = 30
)

// SearchHandler serves GET /v1/search.
type SearchHandler struct {
	repo *repository.SearchRepo
}

func NewSearchHandler(repo *repository.SearchRepo) *SearchHandler {
	return &SearchHandler{repo: repo}
}

// Search returns matches grouped by kind. q is required; limit caps each group.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}
	if len([]rune(q)) > searchMaxQueryLen {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("q must be at most %d characters", searchMaxQueryLen))
		return
	}
	limit := searchDefaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > searchMaxLimit {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("limit must be a whole number from 1 to %d", searchMaxLimit))
			return
		}
		limit = n
	}

	results, err := h.repo.Search(r.Context(), q, limit)
	if err != nil {
		log.Error().Err(err).Msg("search failed")
		writeError(w, http.StatusInternalServerError, "failed to search")
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// RaceHandler serves county race data.
type RaceHandler struct {
	repo *repository.RaceRepo
}

func NewRaceHandler(repo *repository.RaceRepo) *RaceHandler {
	return &RaceHandler{repo: repo}
}

// GetCountyRace serves GET /v1/counties/{code}/race.
func (h *RaceHandler) GetCountyRace(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	race, ok, err := h.repo.CountyRace(r.Context(), code)
	if err != nil {
		log.Error().Err(err).Str("county", code).Msg("county race failed")
		writeError(w, http.StatusInternalServerError, "failed to get county race")
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "county not found")
		return
	}
	writeJSON(w, http.StatusOK, race)
}

// GetProfile serves GET /v1/politicians/{slug}/profile.
func (h *PoliticianHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	profile, err := h.svc.GetProfile(r.Context(), id, chi.URLParam(r, "slug"))
	if err != nil {
		log.Error().Err(err).Msg("politician profile failed")
		writeError(w, http.StatusInternalServerError, "failed to get profile")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// GetPartyProfile serves GET /v1/parties/{slug}/profile.
func (h *PartyHandler) GetPartyProfile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	profile, err := h.repo.GetProfile(r.Context(), slug)
	if err != nil {
		log.Error().Err(err).Str("party", slug).Msg("party profile failed")
		writeError(w, http.StatusInternalServerError, "failed to get party profile")
		return
	}
	if profile == nil {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// GetSummary serves GET /v1/polls/summary?race=... with the recent-polls average for that race.
func (h *PollHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	race := r.URL.Query().Get("race")
	if race == "" {
		race = "president"
	}
	if !containsRace(race) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("race must be one of %v", models.PollRaces))
		return
	}
	summary, err := h.repo.Summary(r.Context(), race, pollSummaryWindow)
	if err != nil {
		log.Error().Err(err).Msg("poll summary failed")
		writeError(w, http.StatusInternalServerError, "failed to get poll summary")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func containsRace(race string) bool {
	for _, r := range models.PollRaces {
		if r == race {
			return true
		}
	}
	return false
}
