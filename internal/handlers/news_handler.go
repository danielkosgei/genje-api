package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type NewsHandler struct {
	repo *repository.NewsRepo
}

func NewNewsHandler(repo *repository.NewsRepo) *NewsHandler {
	return &NewsHandler{repo: repo}
}

func (h *NewsHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	q := r.URL.Query()

	filter := models.NewsFilter{
		Limit:  limit,
		Offset: offset,
	}

	if v := q.Get("source_id"); v != "" {
		id, err := parseUUID(v)
		if err == nil {
			filter.SourceID = &id
		}
	}
	if v := q.Get("politician_id"); v != "" {
		id, err := parseUUID(v)
		if err == nil {
			filter.PoliticianID = &id
		}
	}
	if q.Get("election_related") == "true" {
		t := true
		filter.ElectionRelated = &t
	}

	articles, total, err := h.repo.ListArticles(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list articles")
		return
	}
	if articles == nil {
		articles = []models.NewsArticle{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(articles, total, limit, offset))
}

func (h *NewsHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}

	article, err := h.repo.GetArticleByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get article")
		return
	}
	if article == nil {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}
	writeJSON(w, http.StatusOK, article)
}

func (h *NewsHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.repo.ListDataSources(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sources")
		return
	}
	if sources == nil {
		sources = []models.Source{}
	}
	writeJSON(w, http.StatusOK, sources)
}
