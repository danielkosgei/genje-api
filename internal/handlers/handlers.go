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
	engagementRepo    *repository.EngagementRepository
	aggregatorService *services.AggregatorService
	summarizerService *services.SummarizerService
	trendingService   *services.TrendingService
}

func New(articleRepo *repository.ArticleRepository, sourceRepo *repository.SourceRepository,
	engagementRepo *repository.EngagementRepository, aggregatorService *services.AggregatorService,
	summarizerService *services.SummarizerService, trendingService *services.TrendingService) *Handler {
	return &Handler{
		articleRepo:       articleRepo,
		sourceRepo:        sourceRepo,
		engagementRepo:    engagementRepo,
		aggregatorService: aggregatorService,
		summarizerService: summarizerService,
		trendingService:   trendingService,
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

func (h *Handler) respondValidationError(w http.ResponseWriter, message, details string) {
	h.respondError(w, http.StatusBadRequest, models.ErrCodeValidation, message, details)
}

func (h *Handler) respondNotFound(w http.ResponseWriter, resource string) {
	h.respondError(w, http.StatusNotFound, models.ErrCodeNotFound,
		fmt.Sprintf("%s not found", resource), "")
}

func (h *Handler) respondInternalError(w http.ResponseWriter, message string) {
	h.respondError(w, http.StatusInternalServerError, models.ErrCodeInternal,
		"Internal server error", message)
}

// Enhanced API endpoints

// GetAPIInfo returns API information and available endpoints
func (h *Handler) GetAPIInfo(w http.ResponseWriter, r *http.Request) {
	response := models.APIInfoResponse{
		Name:        "Genje News API",
		Description: "Kenyan news aggregation service providing access to articles from multiple sources",
		Version:     "1.0.0",
		Endpoints: []string{
			"GET /health",
			"GET /",
			"GET /v1/openapi.json",
			"GET /v1/schema", 
			"GET /v1/articles",
			"GET /v1/articles/{id}",
			"POST /v1/articles/{id}/summarize",
			"GET /v1/articles/recent",
			"GET /v1/articles/feed",
			"GET /v1/articles/search",
			"GET /v1/articles/trending",
			"GET /v1/articles/by-source/{sourceId}",
			"GET /v1/articles/by-category/{category}",
			"GET /v1/sources",
			"GET /v1/sources/{id}",
			"POST /v1/sources",
			"PUT /v1/sources/{id}",
			"DELETE /v1/sources/{id}",
			"POST /v1/sources/{id}/refresh",
			"GET /v1/categories",
			"GET /v1/stats",
			"GET /v1/stats/sources",
			"GET /v1/stats/categories",
			"GET /v1/stats/timeline",
			"GET /v1/trends",
			"GET /v1/status",
			"POST /v1/refresh",
		},
		LastUpdated: time.Now(),
	}
	h.respondJSON(w, http.StatusOK, response)
}

// GetGlobalStats returns global statistics
func (h *Handler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.articleRepo.GetGlobalStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get global stats", err.Error())
		return
	}

	// Get total sources
	sourcesCount, err := h.sourceRepo.GetSourcesCount()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get sources count", err.Error())
		return
	}
	stats.TotalSources = sourcesCount

	h.respondSuccess(w, stats)
}

// GetSourceStats returns statistics per source
func (h *Handler) GetSourceStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.articleRepo.GetSourceStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get source stats", err.Error())
		return
	}

	h.respondSuccess(w, stats)
}

// GetCategoryStats returns statistics per category
func (h *Handler) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.articleRepo.GetCategoryStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get category stats", err.Error())
		return
	}

	h.respondSuccess(w, stats)
}

// GetTimelineStats returns article count over time
func (h *Handler) GetTimelineStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30 // default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.articleRepo.GetTimelineStats(days)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get timeline stats", err.Error())
		return
	}

	h.respondSuccess(w, stats)
}

// GetRecentArticles returns recent articles
func (h *Handler) GetRecentArticles(w http.ResponseWriter, r *http.Request) {
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // default
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 168 { // max 1 week
			hours = h
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	articles, err := h.articleRepo.GetRecentArticles(hours, limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get recent articles", err.Error())
		return
	}

	h.respondSuccess(w, articles)
}

// GetSystemStatus returns system status
func (h *Handler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.articleRepo.GetSystemStatus()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get system status", err.Error())
		return
	}

	// Get active sources count
	sourcesCount, err := h.sourceRepo.GetSourcesCount()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get sources count", err.Error())
		return
	}
	status.ActiveSources = sourcesCount

	h.respondSuccess(w, status)
}

// Source management endpoints

// CreateSource creates a new news source
func (h *Handler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondValidationError(w, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := h.validateCreateSourceRequest(req); err != nil {
		h.respondValidationError(w, "Validation failed", err.Error())
		return
	}

	source := &models.NewsSource{
		Name:     req.Name,
		URL:      req.URL,
		FeedURL:  req.FeedURL,
		Category: req.Category,
		Active:   req.Active,
	}

	if err := h.sourceRepo.CreateSource(source); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create source", err.Error())
		return
	}

	h.respondCreated(w, source)
}

// GetSource returns a specific source
func (h *Handler) GetSource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid source ID", "")
		return
	}

	source, err := h.sourceRepo.GetSourceByID(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get source", err.Error())
		return
	}

	if source == nil {
		h.respondError(w, http.StatusNotFound, "Source not found", "")
		return
	}

	h.respondSuccess(w, source)
}

// UpdateSource updates an existing source
func (h *Handler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid source ID", "")
		return
	}

	var req models.UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.sourceRepo.UpdateSource(id, req); err != nil {
		if err.Error() == "source not found" {
			h.respondError(w, http.StatusNotFound, "Source not found", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to update source", err.Error())
		return
	}

	// Get the updated source to return it
	source, err := h.sourceRepo.GetSourceByID(id)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	h.respondSuccess(w, source)
}

// DeleteSource deletes a source
func (h *Handler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid source ID", "")
		return
	}

	if err := h.sourceRepo.DeleteSource(id); err != nil {
		if err.Error() == "source not found" {
			h.respondError(w, http.StatusNotFound, "Source not found", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to delete source", err.Error())
		return
	}

	// Return 204 No Content for successful deletion (RESTful pattern)
	w.WriteHeader(http.StatusNoContent)
}

// RefreshSource refreshes a single source
func (h *Handler) RefreshSource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid source ID", "")
		return
	}

	source, err := h.sourceRepo.RefreshSingleSource(id)
	if err != nil {
		if err.Error() == "source not found" {
			h.respondError(w, http.StatusNotFound, "Source not found", "")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to refresh source", err.Error())
		return
	}

	// Trigger refresh for this source (simplified - in real implementation you'd pass the source to aggregator)
	go func() {
		// This would need to be implemented in the aggregator service
		// _ = h.aggregatorService.AggregateFromSource(r.Context(), *source)
	}()

	response := map[string]interface{}{
		"success": true,
		"message": "Source refresh started",
		"data":    source,
	}
	h.respondJSON(w, http.StatusOK, response)
}

// GetAllSources returns all sources (including inactive ones)
func (h *Handler) GetAllSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.sourceRepo.GetAllSources()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get all sources", err.Error())
		return
	}

	h.respondSuccess(w, sources)
}

// New handler methods for missing endpoints

// GetOpenAPISpec returns OpenAPI specification
func (h *Handler) GetOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := models.OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: models.OpenAPIInfo{
			Title:       "Genje News API",
			Description: "Kenyan news aggregation service providing access to articles from multiple sources",
			Version:     "1.0.0",
			Contact: models.OpenAPIContact{
				Name:  "Genje API Team",
				Email: "support@genje.co.ke",
				URL:   "https://api.genje.co.ke",
			},
		},
		Paths: map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
						},
					},
				},
			},
			"/v1/articles": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get articles",
					"parameters": []interface{}{
						map[string]interface{}{
							"name":     "page",
							"in":       "query",
							"required": false,
							"schema":   map[string]interface{}{"type": "integer"},
						},
						map[string]interface{}{
							"name":     "limit",
							"in":       "query",
							"required": false,
							"schema":   map[string]interface{}{"type": "integer"},
						},
						map[string]interface{}{
							"name":     "category",
							"in":       "query",
							"required": false,
							"schema":   map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response with articles",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ArticlesResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad request - invalid parameters",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Article by ID",
					"description": "Retrieve a specific article by its unique identifier",
					"tags":        []string{"Articles"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Article found",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ArticleResponse",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Article not found",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/{id}/summary": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Article Summary (NLP-Powered)",
					"description": "Get or generate intelligent article summary using advanced NLP techniques including TF-IDF analysis, entity recognition, and multi-criteria scoring",
					"tags":        []string{"Articles", "NLP"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Article summary generated successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SummaryResponse",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Article not found",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Generate New Article Summary",
					"description": "Force generation of a new article summary (same as GET but explicit action)",
					"tags":        []string{"Articles", "NLP"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "New summary generated",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SummaryResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/search": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Search Articles",
					"description": "Full-text search with relevance ranking and advanced filtering",
					"tags":        []string{"Articles", "Search"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "q",
							"in":          "query",
							"description": "Search query (required)",
							"required":    true,
							"schema":      map[string]interface{}{"type": "string", "minLength": 2},
						},
						map[string]interface{}{
							"name":        "category",
							"in":          "query",
							"description": "Filter by category",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						map[string]interface{}{
							"name":        "sort_by",
							"in":          "query",
							"description": "Sort by field",
							"required":    false,
							"schema": map[string]interface{}{
								"type":    "string",
								"enum":    []string{"relevance", "date", "source"},
								"default": "relevance",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Search results",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ArticlesResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/sources": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List News Sources",
					"description": "Get all active news sources",
					"tags":        []string{"Sources"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of active sources",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SourcesResponse",
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create News Source",
					"description": "Add a new news source to the system",
					"tags":        []string{"Sources"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/CreateSourceRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Source created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SourceResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Validation error",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/stats": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Global Statistics",
					"description": "Get comprehensive statistics about articles, sources, and categories",
					"tags":        []string{"Statistics"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Global statistics",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/StatsResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/{id}/engagement": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Article Engagement Metrics",
					"description": "Retrieve engagement metrics (views, shares, comments, likes) for a specific article",
					"tags":        []string{"Articles", "Engagement"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Engagement metrics retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/EngagementResponse",
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Track Article Engagement",
					"description": "Record an engagement event (view, share, comment, like) for an article",
					"tags":        []string{"Articles", "Engagement"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/EngagementRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Engagement tracked successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SuccessResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/trending/advanced": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Advanced Trending Articles",
					"description": "Get trending articles using sophisticated 5-factor algorithm (engagement, velocity, authority, content, recency)",
					"tags":        []string{"Articles", "Trending", "Advanced"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "limit",
							"in":          "query",
							"description": "Number of articles to return (default: 20, max: 100)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 100, "default": 20},
						},
						map[string]interface{}{
							"name":        "window",
							"in":          "query",
							"description": "Time window for trending analysis",
							"required":    false,
							"schema": map[string]interface{}{
								"type":    "string",
								"enum":    []string{"1h", "6h", "12h", "24h", "7d"},
								"default": "24h",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Advanced trending articles with detailed scoring",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/AdvancedTrendingResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/sources/{name}/authority": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Source Authority Metrics",
					"description": "Retrieve authority, credibility, and reach scores for a news source",
					"tags":        []string{"Sources", "Authority"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "name",
							"in":          "path",
							"description": "Source name",
							"required":    true,
							"schema":      map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Source authority metrics",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SourceAuthorityResponse",
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Update Source Authority",
					"description": "Manually trigger recalculation of source authority metrics",
					"tags":        []string{"Sources", "Authority"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "name",
							"in":          "path",
							"description": "Source name",
							"required":    true,
							"schema":      map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Authority metrics updated",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/SourceAuthorityResponse",
									},
								},
							},
						},
					},
				},
			},
		},
		Components: map[string]interface{}{
			"schemas": map[string]interface{}{
				"Article": map[string]interface{}{
					"type":        "object",
					"description": "News article with comprehensive metadata",
					"properties": map[string]interface{}{
						"id":           map[string]interface{}{"type": "integer", "example": 123, "description": "Unique article identifier"},
						"title":        map[string]interface{}{"type": "string", "example": "Kenya's Economic Growth Outlook for 2025", "description": "Article headline"},
						"content":      map[string]interface{}{"type": "string", "description": "Full article content (HTML)"},
						"summary":      map[string]interface{}{"type": "string", "example": "Economic experts predict steady growth driven by infrastructure investments.", "description": "AI-generated summary using NLP techniques"},
						"url":          map[string]interface{}{"type": "string", "format": "uri", "example": "https://standardmedia.co.ke/business/article/2025/01/15/kenya-economic-growth", "description": "Original article URL"},
						"author":       map[string]interface{}{"type": "string", "example": "Jane Doe", "description": "Article author"},
						"source":       map[string]interface{}{"type": "string", "example": "Standard Business", "description": "News source name"},
						"published_at": map[string]interface{}{"type": "string", "format": "date-time", "example": "2025-01-15T10:30:00Z", "description": "Original publication date"},
						"created_at":   map[string]interface{}{"type": "string", "format": "date-time", "example": "2025-01-15T10:35:00Z", "description": "Date added to our system"},
						"category":     map[string]interface{}{"type": "string", "example": "business", "description": "Article category", "enum": []string{"news", "sports", "business", "politics", "technology", "entertainment", "health", "world", "opinion", "general", "kiswahili", "diaspora"}},
						"image_url":    map[string]interface{}{"type": "string", "format": "uri", "example": "https://standardmedia.co.ke/images/business/economic-growth.jpg", "description": "Article featured image"},
					},
					"required": []string{"id", "title", "url", "source", "published_at", "created_at", "category"},
				},
				"APIResponse": map[string]interface{}{
					"type":        "object",
					"description": "Standardized API response wrapper",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true, "description": "Indicates if the request was successful"},
						"data":    map[string]interface{}{"description": "Response data (varies by endpoint)"},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp":  map[string]interface{}{"type": "string", "format": "date-time", "description": "Response generation timestamp"},
								"pagination": map[string]interface{}{"$ref": "#/components/schemas/PaginationMeta"},
								"request_id": map[string]interface{}{"type": "string", "description": "Unique request identifier for debugging"},
							},
						},
					},
					"required": []string{"success"},
				},
				"PaginationMeta": map[string]interface{}{
					"type":        "object",
					"description": "Enhanced pagination information",
					"properties": map[string]interface{}{
						"page":        map[string]interface{}{"type": "integer", "example": 1, "description": "Current page number"},
						"limit":       map[string]interface{}{"type": "integer", "example": 20, "description": "Items per page"},
						"total":       map[string]interface{}{"type": "integer", "example": 1250, "description": "Total number of items"},
						"total_pages": map[string]interface{}{"type": "integer", "example": 63, "description": "Total number of pages"},
						"has_next":    map[string]interface{}{"type": "boolean", "example": true, "description": "Whether there are more pages"},
						"has_prev":    map[string]interface{}{"type": "boolean", "example": false, "description": "Whether there are previous pages"},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type":        "object",
					"description": "Standardized error response",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": false},
						"error": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":    map[string]interface{}{"type": "string", "example": "VALIDATION_ERROR", "description": "Error code for programmatic handling"},
								"message": map[string]interface{}{"type": "string", "example": "Invalid query parameters", "description": "Human-readable error message"},
								"details": map[string]interface{}{"type": "string", "example": "Page parameter must be a positive integer", "description": "Additional error details"},
							},
						},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp":  map[string]interface{}{"type": "string", "format": "date-time"},
								"request_id": map[string]interface{}{"type": "string"},
							},
						},
					},
				},
				"SummaryResponse": map[string]interface{}{
					"type":        "object",
					"description": "NLP-powered article summary response",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"summary": map[string]interface{}{
									"type":        "string",
									"example":     "President William Ruto has pledged to expand youth employment in both the Climate Worx and affordable housing programmes, aiming to double current figures within the next three months.",
									"description": "AI-generated summary using TF-IDF analysis, entity recognition, and multi-criteria scoring",
								},
							},
						},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
							},
						},
					},
				},
			},
		},
	}

	h.respondJSON(w, http.StatusOK, spec)
}

// GetAPISchema returns API schema information
func (h *Handler) GetAPISchema(w http.ResponseWriter, r *http.Request) {
	schema := models.APISchema{
		Models: map[string]interface{}{
			"Article": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":           map[string]interface{}{"type": "integer"},
					"title":        map[string]interface{}{"type": "string"},
					"content":      map[string]interface{}{"type": "string"},
					"url":          map[string]interface{}{"type": "string"},
					"author":       map[string]interface{}{"type": "string"},
					"source":       map[string]interface{}{"type": "string"},
					"published_at": map[string]interface{}{"type": "string", "format": "date-time"},
					"created_at":   map[string]interface{}{"type": "string", "format": "date-time"},
					"category":     map[string]interface{}{"type": "string"},
				},
			},
			"NewsSource": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":       map[string]interface{}{"type": "integer"},
					"name":     map[string]interface{}{"type": "string"},
					"url":      map[string]interface{}{"type": "string"},
					"feed_url": map[string]interface{}{"type": "string"},
					"category": map[string]interface{}{"type": "string"},
					"active":   map[string]interface{}{"type": "boolean"},
				},
			},
		},
		Endpoints: []models.EndpointSchema{
			{
				Path:        "/v1/articles",
				Method:      "GET",
				Description: "Get paginated list of articles",
				Parameters: []models.ParameterSchema{
					{Name: "page", In: "query", Required: false, Type: "integer", Description: "Page number (default: 1)"},
					{Name: "limit", In: "query", Required: false, Type: "integer", Description: "Items per page (default: 20, max: 100)"},
					{Name: "category", In: "query", Required: false, Type: "string", Description: "Filter by category"},
					{Name: "source", In: "query", Required: false, Type: "string", Description: "Filter by source"},
					{Name: "search", In: "query", Required: false, Type: "string", Description: "Search in title and content"},
				},
				Response: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"articles":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/models/Article"}},
						"pagination": map[string]interface{}{"type": "object"},
					},
				},
			},
			{
				Path:        "/v1/articles/{id}",
				Method:      "GET",
				Description: "Get single article by ID",
				Parameters: []models.ParameterSchema{
					{Name: "id", In: "path", Required: true, Type: "integer", Description: "Article ID"},
				},
				Response: map[string]interface{}{
					"$ref": "#/models/Article",
				},
			},
		},
	}

	h.respondJSON(w, http.StatusOK, schema)
}

// SearchArticles performs full-text search
func (h *Handler) SearchArticles(w http.ResponseWriter, r *http.Request) {
	filters := models.SearchFilters{
		Query:     r.URL.Query().Get("q"),
		Category:  r.URL.Query().Get("category"),
		Source:    r.URL.Query().Get("source"),
		From:      r.URL.Query().Get("from"),
		To:        r.URL.Query().Get("to"),
		Page:      1,
		Limit:     20,
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
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

	if filters.Query == "" {
		h.respondError(w, http.StatusBadRequest, "Query parameter 'q' is required", "")
		return
	}

	articles, total, err := h.articleRepo.SearchArticles(filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to search articles", err.Error())
		return
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit
	pagination := models.PaginationMeta{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    filters.Page < totalPages,
		HasPrev:    filters.Page > 1,
	}

	response := models.NewPaginatedResponse(articles, pagination)
	h.respondJSON(w, http.StatusOK, response)
}

// GetArticlesFeed returns cursor-based paginated feed
func (h *Handler) GetArticlesFeed(w http.ResponseWriter, r *http.Request) {
	req := models.FeedRequest{
		Cursor:    r.URL.Query().Get("cursor"),
		Limit:     20,
		Category:  r.URL.Query().Get("category"),
		Source:    r.URL.Query().Get("source"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	feed, err := h.articleRepo.GetArticlesFeed(req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get articles feed", err.Error())
		return
	}

	h.respondSuccess(w, feed)
}

// GetTrendingArticles returns trending articles using the advanced algorithm
func (h *Handler) GetTrendingArticles(w http.ResponseWriter, r *http.Request) {
	limit := 20
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h"
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Use the advanced trending algorithm
	trending, err := h.trendingService.GetAdvancedTrendingArticles(limit, timeWindow)
	if err != nil {
		// Fallback to simple trending if advanced fails
		articles, fallbackErr := h.articleRepo.GetTrendingArticles(limit, timeWindow)
		if fallbackErr != nil {
			h.respondInternalError(w, err.Error())
			return
		}
		h.respondSuccess(w, articles)
		return
	}

	response := map[string]interface{}{
		"articles":    trending,
		"time_window": timeWindow,
		"algorithm":   "5-factor advanced trending",
	}

	h.respondSuccess(w, response)
}

// GetArticlesBySource returns articles from a specific source
func (h *Handler) GetArticlesBySource(w http.ResponseWriter, r *http.Request) {
	sourceIDStr := chi.URLParam(r, "sourceId")
	sourceID, err := strconv.Atoi(sourceIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid source ID", "")
		return
	}

	filters, err := h.parseArticleFilters(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	articles, total, err := h.articleRepo.GetArticlesBySource(sourceID, filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get articles by source", err.Error())
		return
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit
	pagination := models.PaginationMeta{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    filters.Page < totalPages,
		HasPrev:    filters.Page > 1,
	}

	response := models.NewPaginatedResponse(articles, pagination)
	h.respondJSON(w, http.StatusOK, response)
}

// GetArticlesByCategory returns articles from a specific category
func (h *Handler) GetArticlesByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		h.respondError(w, http.StatusBadRequest, "Category parameter is required", "")
		return
	}

	filters, err := h.parseArticleFilters(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	articles, total, err := h.articleRepo.GetArticlesByCategory(category, filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get articles by category", err.Error())
		return
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit
	pagination := models.PaginationMeta{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    filters.Page < totalPages,
		HasPrev:    filters.Page > 1,
	}

	response := models.NewPaginatedResponse(articles, pagination)
	h.respondJSON(w, http.StatusOK, response)
}

// GetTrends returns trending topics/keywords
func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	limit := 10
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h"
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	topics, err := h.articleRepo.GetTrendingTopics(limit, timeWindow)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get trending topics", err.Error())
		return
	}

	h.respondSuccess(w, topics)
}

// validateCreateSourceRequest validates the create source request
func (h *Handler) validateCreateSourceRequest(req models.CreateSourceRequest) error {
	validator := validation.New()

	// Required fields
	validator.Required("name", req.Name)
	validator.Required("feed_url", req.FeedURL)

	// Length validations
	validator.MinLength("name", req.Name, 2)
	validator.MaxLength("name", req.Name, 100)

	// URL validations
	validator.URL("url", req.URL)
	validator.URL("feed_url", req.FeedURL)

	// Category validation
	allowedCategories := []string{"news", "sports", "business", "politics", "technology", "entertainment", "health", "world", "opinion", "general"}
	validator.In("category", req.Category, allowedCategories)

	if validator.HasErrors() {
		return validator.Errors()
	}

	return nil
}
