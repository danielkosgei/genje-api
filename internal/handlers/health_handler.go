package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type HealthHandler struct {
	pool    *pgxpool.Pool
	startAt time.Time
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool, startAt: time.Now()}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbOK := true

	if err := h.pool.Ping(ctx); err != nil {
		dbOK = false
	}

	status := "healthy"
	code := http.StatusOK
	if !dbOK {
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, map[string]interface{}{
		"status":   status,
		"database": dbOK,
		"uptime":   time.Since(h.startAt).String(),
	})
}

func (h *HealthHandler) APIInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.APIInfo{
		Name:        "Jalada",
		Version:     "1.0.0",
		Description: "Kenya 2027 Election Tracker API — neutral, comprehensive, sourced.",
	})
}
