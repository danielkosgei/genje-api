package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type PartyHandler struct {
	repo *repository.PartyRepo
}

func NewPartyHandler(repo *repository.PartyRepo) *PartyHandler {
	return &PartyHandler{repo: repo}
}

func (h *PartyHandler) ListParties(w http.ResponseWriter, r *http.Request) {
	parties, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list parties")
		return
	}
	if parties == nil {
		parties = []models.PoliticalParty{}
	}
	writeJSON(w, http.StatusOK, parties)
}

func (h *PartyHandler) GetParty(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	party, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get party")
		return
	}
	if party == nil {
		writeError(w, http.StatusNotFound, "party not found")
		return
	}

	members, _ := h.repo.GetMembers(r.Context(), slug)
	if members == nil {
		members = []models.PoliticianSummary{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"party":   party,
		"members": members,
	})
}

func (h *PartyHandler) ListCoalitions(w http.ResponseWriter, r *http.Request) {
	coalitions, err := h.repo.ListCoalitions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list coalitions")
		return
	}
	if coalitions == nil {
		coalitions = []models.Coalition{}
	}
	writeJSON(w, http.StatusOK, coalitions)
}

func (h *PartyHandler) GetCoalition(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	coalition, err := h.repo.GetCoalitionBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get coalition")
		return
	}
	if coalition == nil {
		writeError(w, http.StatusNotFound, "coalition not found")
		return
	}
	writeJSON(w, http.StatusOK, coalition)
}
