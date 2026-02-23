package handlers

import (
	"net/http"

	"jalada/internal/models"
	"jalada/internal/services"
)

type TimelineHandler struct {
	svc *services.TimelineService
}

func NewTimelineHandler(svc *services.TimelineService) *TimelineHandler {
	return &TimelineHandler{svc: svc}
}

func (h *TimelineHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.GetElectionTimeline(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get timeline")
		return
	}
	if events == nil {
		events = []models.TimelineEvent{}
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *TimelineHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	events, total, err := h.svc.ListEvents(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	if events == nil {
		events = []models.Event{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(events, total, limit, offset))
}
