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
	
	response := models.StatsResponse{
		Success: true,
		Data:    stats,
	}
	response.Meta.GeneratedAt = time.Now()
	
	h.respondJSON(w, http.StatusOK, response)
}

// GetSourceStats returns statistics per source
func (h *Handler) GetSourceStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.articleRepo.GetSourceStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get source stats", err.Error())
		return
	}
	
	response := models.SourceStatsResponse{
		Success: true,
		Data:    stats,
	}
	response.Meta.GeneratedAt = time.Now()
	
	h.respondJSON(w, http.StatusOK, response)
}

// GetCategoryStats returns statistics per category
func (h *Handler) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.articleRepo.GetCategoryStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get category stats", err.Error())
		return
	}
	
	response := models.CategoryStatsResponse{
		Success: true,
		Data:    stats,
	}
	response.Meta.GeneratedAt = time.Now()
	
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := models.TimelineStatsResponse{
		Success: true,
		Data:    stats,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Days = days
	
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := models.RecentArticlesResponse{
		Success: true,
		Data:    articles,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Hours = hours
	response.Meta.Total = len(articles)
	
	h.respondJSON(w, http.StatusOK, response)
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
	
	h.respondJSON(w, http.StatusOK, status)
}

// Source management endpoints

// CreateSource creates a new news source
func (h *Handler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
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
	
	response := map[string]interface{}{
		"success": true,
		"message": "Source created successfully",
		"data":    source,
	}
	h.respondJSON(w, http.StatusCreated, response)
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
	
	response := map[string]interface{}{
		"success": true,
		"data":    source,
	}
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := map[string]interface{}{
		"success": true,
		"message": "Source updated successfully",
	}
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := map[string]interface{}{
		"success": true,
		"message": "Source deleted successfully",
	}
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := map[string]interface{}{
		"success": true,
		"data":    sources,
		"meta": map[string]interface{}{
			"total":        len(sources),
			"generated_at": time.Now(),
		},
	}
	h.respondJSON(w, http.StatusOK, response)
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
				Name:  "Genje Team",
				Email: "support@genje.com",
				URL:   "https://genje.com",
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
							"description": "List of articles",
						},
					},
				},
			},
		},
		Components: map[string]interface{}{
			"schemas": map[string]interface{}{
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
	
	response := models.SearchResponse{
		Success: true,
		Data:    articles,
	}
	response.Meta.Pagination = models.PaginationResponse{
		Page:  filters.Page,
		Limit: filters.Limit,
		Total: total,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Query = filters.Query
	response.Meta.Filters = filters
	
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
	
	response := models.FeedResponse{
		Success: true,
		Data:    feed,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Filters = req
	
	h.respondJSON(w, http.StatusOK, response)
}

// GetTrendingArticles returns trending articles
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
	
	articles, err := h.articleRepo.GetTrendingArticles(limit, timeWindow)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get trending articles", err.Error())
		return
	}
	
	response := models.TrendingResponse{
		Success: true,
		Data:    articles,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Algorithm = "recent_engagement"
	response.Meta.TimeWindow = timeWindow
	
	h.respondJSON(w, http.StatusOK, response)
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
	
	response := models.EnhancedArticlesResponse{
		Success: true,
		Data:    articles,
	}
	response.Meta.Pagination = models.PaginationResponse{
		Page:  filters.Page,
		Limit: filters.Limit,
		Total: total,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Filters = filters
	
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
	
	response := models.EnhancedArticlesResponse{
		Success: true,
		Data:    articles,
	}
	response.Meta.Pagination = models.PaginationResponse{
		Page:  filters.Page,
		Limit: filters.Limit,
		Total: total,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.Filters = filters
	
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
	
	response := models.TrendsResponse{
		Success: true,
		Data:    topics,
	}
	response.Meta.GeneratedAt = time.Now()
	response.Meta.TimeWindow = timeWindow
	response.Meta.Algorithm = "category_frequency"
	
	h.respondJSON(w, http.StatusOK, response)
} 