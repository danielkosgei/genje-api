package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"genje-api/internal/models"
	"genje-api/internal/repository"
	"genje-api/internal/services"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	articleRepo       *repository.ArticleRepository
	sourceRepo        *repository.SourceRepository
	aggregatorService *services.AggregatorService
	summarizerService *services.SummarizerService
}

func New(articleRepo *repository.ArticleRepository, sourceRepo *repository.SourceRepository, 
	aggregatorService *services.AggregatorService, summarizerService *services.SummarizerService) *Handler {
	return &Handler{
		articleRepo:       articleRepo,
		sourceRepo:        sourceRepo,
		aggregatorService: aggregatorService,
		summarizerService: summarizerService,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
	filters, err := h.parseArticleFilters(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	articles, total, err := h.articleRepo.GetArticles(filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get articles", err.Error())
		return
	}

	response := models.ArticlesResponse{
		Articles: articles,
		Pagination: models.PaginationResponse{
			Page:  filters.Page,
			Limit: filters.Limit,
			Total: total,
		},
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid article ID", "")
		return
	}

	article, err := h.articleRepo.GetArticleByID(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get article", err.Error())
		return
	}

	if article == nil {
		h.respondError(w, http.StatusNotFound, "Article not found", "")
		return
	}

	h.respondJSON(w, http.StatusOK, article)
}

func (h *Handler) SummarizeArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid article ID", "")
		return
	}

	summary, err := h.summarizerService.SummarizeArticle(id)
	if err != nil {
		if err.Error() == "article not found" {
			h.respondError(w, http.StatusNotFound, "Article not found", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to summarize article", err.Error())
		return
	}

	response := map[string]string{"summary": summary}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.sourceRepo.GetActiveSources()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get sources", err.Error())
		return
	}

	response := models.SourcesResponse{Sources: sources}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.articleRepo.GetCategories()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}

	response := models.CategoriesResponse{Categories: categories}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) RefreshNews(w http.ResponseWriter, r *http.Request) {
	go func() {
		_ = h.aggregatorService.AggregateNews(r.Context()) // Error logged in service
	}()
	
	response := map[string]string{"message": "News refresh started"}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) parseArticleFilters(r *http.Request) (models.ArticleFilters, error) {
	filters := models.ArticleFilters{
		Page:     1,
		Limit:    20,
		Category: r.URL.Query().Get("category"),
		Source:   r.URL.Query().Get("source"),
		Search:   r.URL.Query().Get("search"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	return filters, nil
}

func (h *Handler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't change response as headers are already written
		// In a real application, you might want to use a proper logger here
		return
	}
}

func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message, details string) {
	response := models.ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Details: details,
	}
	h.respondJSON(w, statusCode, response)
} 