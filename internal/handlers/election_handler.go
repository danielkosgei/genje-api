package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jalada/internal/models"
	"jalada/internal/services"
)

type ElectionHandler struct {
	svc *services.ElectionService
}

func NewElectionHandler(svc *services.ElectionService) *ElectionHandler {
	return &ElectionHandler{svc: svc}
}

func (h *ElectionHandler) List(w http.ResponseWriter, r *http.Request) {
	elections, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list elections")
		return
	}
	if elections == nil {
		elections = []models.Election{}
	}
	writeJSON(w, http.StatusOK, elections)
}

func (h *ElectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid election id")
		return
	}

	election, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get election")
		return
	}
	if election == nil {
		writeError(w, http.StatusNotFound, "election not found")
		return
	}
	writeJSON(w, http.StatusOK, election)
}

func (h *ElectionHandler) GetCandidates(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid election id")
		return
	}

	candidates, err := h.svc.GetCandidates(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get candidates")
		return
	}
	if candidates == nil {
		candidates = []models.CandidacyDetail{}
	}
	writeJSON(w, http.StatusOK, candidates)
}

func (h *ElectionHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid election id")
		return
	}

	results, err := h.svc.GetResults(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get results")
		return
	}
	if results == nil {
		results = []models.ResultSummary{}
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *ElectionHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid election id")
		return
	}

	timeline, err := h.svc.GetTimeline(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get timeline")
		return
	}
	if timeline == nil {
		timeline = []models.TimelineEvent{}
	}
	writeJSON(w, http.StatusOK, timeline)
}
