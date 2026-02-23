package handlers

import (
	"net/http"

	"jalada/internal/services"
)

type AnalyticsHandler struct {
	svc *services.AnalyticsService
}

func NewAnalyticsHandler(svc *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) Sentiment(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetNationalSentiment(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get sentiment analytics")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) Promises(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetPromiseAnalytics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get promise analytics")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) Integrity(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetIntegrityAnalytics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get integrity analytics")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) Attendance(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetAttendanceAnalytics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get attendance analytics")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) Trending(w http.ResponseWriter, r *http.Request) {
	limit, _ := parsePagination(r)
	data, err := h.svc.GetTrending(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get trending")
		return
	}
	writeJSON(w, http.StatusOK, data)
}
