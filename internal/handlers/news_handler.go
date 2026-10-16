package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"jalada/internal/models"
)

// NewsStore is the data the news endpoints need. The repository satisfies it; tests use a fake.
type NewsStore interface {
	ListArticles(ctx context.Context, f models.NewsFilter) ([]models.NewsArticle, int, error)
	GetArticleByID(ctx context.Context, id uuid.UUID) (*models.NewsArticle, error)
	ListDataSources(ctx context.Context) ([]models.Source, error)
}

type NewsHandler struct {
	store NewsStore
}

func NewNewsHandler(store NewsStore) *NewsHandler {
	return &NewsHandler{store: store}
}

func (h *NewsHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	filter, err := parseNewsFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	articles, total, err := h.store.ListArticles(r.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("list articles failed")
		writeError(w, http.StatusInternalServerError, "failed to list articles")
		return
	}
	if articles == nil {
		articles = []models.NewsArticle{}
	}
	writeJSON(w, http.StatusOK, models.NewPaginatedResponse(articles, total, filter.Limit, filter.Offset))
}

func (h *NewsHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}

	article, err := h.store.GetArticleByID(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("article_id", id.String()).Msg("get article failed")
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
	sources, err := h.store.ListDataSources(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("list sources failed")
		writeError(w, http.StatusInternalServerError, "failed to list sources")
		return
	}
	if sources == nil {
		sources = []models.Source{}
	}
	writeJSON(w, http.StatusOK, sources)
}
