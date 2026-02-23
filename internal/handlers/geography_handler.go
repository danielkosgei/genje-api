package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type GeographyHandler struct {
	repo *repository.GeographyRepo
}

func NewGeographyHandler(repo *repository.GeographyRepo) *GeographyHandler {
	return &GeographyHandler{repo: repo}
}

func (h *GeographyHandler) ListCounties(w http.ResponseWriter, r *http.Request) {
	counties, err := h.repo.ListCounties(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list counties")
		return
	}
	if counties == nil {
		counties = []models.County{}
	}
	writeJSON(w, http.StatusOK, counties)
}

func (h *GeographyHandler) GetCounty(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	county, err := h.repo.GetCountyByCode(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get county")
		return
	}
	if county == nil {
		writeError(w, http.StatusNotFound, "county not found")
		return
	}
	writeJSON(w, http.StatusOK, county)
}

func (h *GeographyHandler) GetConstituencies(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	constituencies, err := h.repo.GetConstituenciesByCounty(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get constituencies")
		return
	}
	if constituencies == nil {
		constituencies = []models.Constituency{}
	}
	writeJSON(w, http.StatusOK, constituencies)
}

func (h *GeographyHandler) GetConstituency(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	constituency, err := h.repo.GetConstituencyByCode(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get constituency")
		return
	}
	if constituency == nil {
		writeError(w, http.StatusNotFound, "constituency not found")
		return
	}
	writeJSON(w, http.StatusOK, constituency)
}

func (h *GeographyHandler) GetWards(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	wards, err := h.repo.GetWardsByConstituency(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get wards")
		return
	}
	if wards == nil {
		wards = []models.Ward{}
	}
	writeJSON(w, http.StatusOK, wards)
}

func (h *GeographyHandler) GetPollingStations(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	stations, err := h.repo.GetPollingStationsByConstituency(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get polling stations")
		return
	}
	if stations == nil {
		stations = []models.PollingStation{}
	}
	writeJSON(w, http.StatusOK, stations)
}

func (h *GeographyHandler) GetCandidates(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	candidates, err := h.repo.GetCandidatesByConstituency(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get candidates")
		return
	}
	if candidates == nil {
		candidates = []models.CandidacyDetail{}
	}
	writeJSON(w, http.StatusOK, candidates)
}
