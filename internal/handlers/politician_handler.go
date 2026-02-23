package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"jalada/internal/models"
	"jalada/internal/services"
)

type PoliticianHandler struct {
	svc *services.PoliticianService
}

func NewPoliticianHandler(svc *services.PoliticianService) *PoliticianHandler {
	return &PoliticianHandler{svc: svc}
}

func (h *PoliticianHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	q := r.URL.Query()

	filter := models.PoliticianFilter{
		Query:  q.Get("q"),
		Limit:  limit,
		Offset: offset,
	}

	if pid := q.Get("party_id"); pid != "" {
		id, err := parseUUID(pid)
		if err == nil {
			filter.PartyID = &id
		}
	}

	politicians, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list politicians")
		return
	}
	if politicians == nil {
		politicians = []models.PoliticianSummary{}
	}

	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(politicians, total, limit, offset))
}

func (h *PoliticianHandler) GetDossier(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	dossier, err := h.svc.GetDossier(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get politician dossier")
		return
	}
	if dossier == nil {
		writeError(w, http.StatusNotFound, "politician not found")
		return
	}
	writeJSON(w, http.StatusOK, dossier)
}

func (h *PoliticianHandler) resolvePoliticianID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	slug := chi.URLParam(r, "slug")
	p, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find politician")
		return uuid.UUID{}, false
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "politician not found")
		return uuid.UUID{}, false
	}
	return p.ID, true
}

func (h *PoliticianHandler) GetNews(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	limit, offset := parsePagination(r)

	articles, total, err := h.svc.GetNews(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get news")
		return
	}
	if articles == nil {
		articles = []models.NewsArticle{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(articles, total, limit, offset))
}

func (h *PoliticianHandler) GetVotingRecord(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	records, err := h.svc.GetVotingRecords(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get voting records")
		return
	}
	if records == nil {
		records = []models.VotingRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *PoliticianHandler) GetCourtCases(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	cases, err := h.svc.GetCourtCases(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get court cases")
		return
	}
	if cases == nil {
		cases = []models.CourtCase{}
	}
	writeJSON(w, http.StatusOK, cases)
}

func (h *PoliticianHandler) GetPromises(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	promises, err := h.svc.GetPromises(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get promises")
		return
	}

	stats, _ := h.svc.GetPromiseStats(r.Context(), id)

	if promises == nil {
		promises = []models.Promise{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"promises": promises,
		"stats":    stats,
	})
}

func (h *PoliticianHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	achievements, err := h.svc.GetAchievements(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get achievements")
		return
	}
	if achievements == nil {
		achievements = []models.Achievement{}
	}
	writeJSON(w, http.StatusOK, achievements)
}

func (h *PoliticianHandler) GetControversies(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	controversies, err := h.svc.GetControversies(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get controversies")
		return
	}
	if controversies == nil {
		controversies = []models.Controversy{}
	}
	writeJSON(w, http.StatusOK, controversies)
}

func (h *PoliticianHandler) GetAffiliations(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	affiliations, err := h.svc.GetAffiliations(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get affiliations")
		return
	}
	if affiliations == nil {
		affiliations = []models.AffiliationDetail{}
	}
	writeJSON(w, http.StatusOK, affiliations)
}

func (h *PoliticianHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	declarations, err := h.svc.GetAssetDeclarations(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get asset declarations")
		return
	}
	if declarations == nil {
		declarations = []models.AssetDeclaration{}
	}
	writeJSON(w, http.StatusOK, declarations)
}

func (h *PoliticianHandler) GetAttendance(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	stats, err := h.svc.GetAttendanceStats(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get attendance")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *PoliticianHandler) GetSentiment(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	limit, _ := parsePagination(r)
	snapshots, err := h.svc.GetSentiment(r.Context(), id, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get sentiment")
		return
	}
	if snapshots == nil {
		snapshots = []models.SentimentSnapshot{}
	}
	writeJSON(w, http.StatusOK, snapshots)
}

func (h *PoliticianHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolvePoliticianID(w, r)
	if !ok {
		return
	}
	limit, offset := parsePagination(r)
	events, total, err := h.svc.GetEvents(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get events")
		return
	}
	if events == nil {
		events = []models.Event{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(events, total, limit, offset))
}
