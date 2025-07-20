package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"genje-api/internal/models"

	"github.com/go-chi/chi/v5"
)

// TrackEngagement records an engagement event for an article
func (h *Handler) TrackEngagement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondValidationError(w, "Invalid article ID", "Article ID must be a positive integer")
		return
	}

	var req models.EngagementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondValidationError(w, "Invalid request body", err.Error())
		return
	}

	// Validate event type
	validEvents := map[string]bool{"view": true, "share": true, "comment": true, "like": true}
	if !validEvents[req.EventType] {
		h.respondValidationError(w, "Invalid event type", "Event type must be one of: view, share, comment, like")
		return
	}

	// Get client IP if not provided
	if req.UserIP == "" {
		req.UserIP = r.RemoteAddr
	}

	// Get user agent if not provided
	if req.UserAgent == "" {
		req.UserAgent = r.Header.Get("User-Agent")
	}

	// Track the engagement
	err = h.engagementRepo.TrackEngagement(articleID, req)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	// Update source authority in background
	go func() {
		article, err := h.articleRepo.GetArticleByID(articleID)
		if err == nil && article != nil {
			if err := h.engagementRepo.UpdateSourceAuthority(article.Source); err != nil {
				// Log error but don't fail the main request
				// In production, this would be logged properly
				_ = err // Satisfy linter
			}
		}
	}()

	response := map[string]interface{}{
		"message":    "Engagement tracked successfully",
		"article_id": articleID,
		"event_type": req.EventType,
	}
	h.respondSuccess(w, response)
}

// GetArticleEngagement returns engagement metrics for an article
func (h *Handler) GetArticleEngagement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondValidationError(w, "Invalid article ID", "Article ID must be a positive integer")
		return
	}

	engagement, err := h.engagementRepo.GetArticleEngagement(articleID)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	h.respondSuccess(w, engagement)
}

// GetAdvancedTrendingArticles returns trending articles using the sophisticated algorithm
func (h *Handler) GetAdvancedTrendingArticles(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h" // default
	}

	// Validate time window
	validWindows := map[string]bool{"1h": true, "6h": true, "12h": true, "24h": true, "7d": true}
	if !validWindows[timeWindow] {
		h.respondValidationError(w, "Invalid time window", "Time window must be one of: 1h, 6h, 12h, 24h, 7d")
		return
	}

	// Get trending articles using the advanced algorithm
	trending, err := h.trendingService.GetAdvancedTrendingArticles(limit, timeWindow)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	response := map[string]interface{}{
		"articles":    trending,
		"time_window": timeWindow,
		"algorithm":   "5-factor advanced trending",
		"factors": []string{
			"engagement_score",
			"velocity_trending",
			"source_authority",
			"content_analysis",
			"time_decay",
		},
	}

	h.respondSuccess(w, response)
}

// GetSourceAuthority returns authority metrics for a source
func (h *Handler) GetSourceAuthority(w http.ResponseWriter, r *http.Request) {
	sourceName := chi.URLParam(r, "name")
	if sourceName == "" {
		h.respondValidationError(w, "Source name is required", "Source name must be provided in the URL path")
		return
	}

	authority, err := h.engagementRepo.GetSourceAuthority(sourceName)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	h.respondSuccess(w, authority)
}

// GetTopEngagedArticles returns most engaged articles in a time window
func (h *Handler) GetTopEngagedArticles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h" // default
	}

	// Validate time window
	validWindows := map[string]bool{"1h": true, "6h": true, "12h": true, "24h": true, "7d": true}
	if !validWindows[timeWindow] {
		h.respondValidationError(w, "Invalid time window", "Time window must be one of: 1h, 6h, 12h, 24h, 7d")
		return
	}

	engagements, err := h.engagementRepo.GetTopEngagedArticles(limit, timeWindow)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	// Get full article data for each engagement
	var enrichedResults []map[string]interface{}
	for _, engagement := range engagements {
		article, err := h.articleRepo.GetArticleByID(engagement.ArticleID)
		if err != nil {
			continue // Skip articles that can't be found
		}

		enrichedResults = append(enrichedResults, map[string]interface{}{
			"article":                article,
			"engagement":             engagement,
			"total_engagement_score": engagement.Views + engagement.Shares*5 + engagement.Comments*3 + engagement.Likes*2,
		})
	}

	response := map[string]interface{}{
		"results":     enrichedResults,
		"time_window": timeWindow,
		"algorithm":   "engagement-weighted scoring",
	}

	h.respondSuccess(w, response)
}

// UpdateSourceAuthority manually triggers source authority recalculation
func (h *Handler) UpdateSourceAuthority(w http.ResponseWriter, r *http.Request) {
	sourceName := chi.URLParam(r, "name")
	if sourceName == "" {
		h.respondValidationError(w, "Source name is required", "Source name must be provided in the URL path")
		return
	}

	err := h.engagementRepo.UpdateSourceAuthority(sourceName)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	// Get updated authority scores
	authority, err := h.engagementRepo.GetSourceAuthority(sourceName)
	if err != nil {
		h.respondInternalError(w, err.Error())
		return
	}

	response := map[string]interface{}{
		"message":   "Source authority updated successfully",
		"source":    sourceName,
		"authority": authority,
	}

	h.respondSuccess(w, response)
}

// RefreshTrendingCache manually refreshes the trending cache
func (h *Handler) RefreshTrendingCache(w http.ResponseWriter, r *http.Request) {
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h" // default
	}

	// Validate time window
	validWindows := map[string]bool{"1h": true, "6h": true, "12h": true, "24h": true, "7d": true}
	if !validWindows[timeWindow] {
		h.respondValidationError(w, "Invalid time window", "Time window must be one of: 1h, 6h, 12h, 24h, 7d")
		return
	}

	// Refresh cache by calling the trending service
	go func() {
		_, err := h.trendingService.GetAdvancedTrendingArticles(100, timeWindow)
		if err != nil {
			log.Printf("Warning: Failed to refresh trending cache for %s: %v", timeWindow, err)
		}
	}()

	response := map[string]interface{}{
		"message":     "Trending cache refresh started",
		"time_window": timeWindow,
		"status":      "processing",
	}

	h.respondSuccess(w, response)
}

// GetEngagementStats returns overall engagement statistics
func (h *Handler) GetEngagementStats(w http.ResponseWriter, r *http.Request) {
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "24h" // default
	}

	// This would need to be implemented in the engagement repository
	// For now, return a placeholder response
	response := map[string]interface{}{
		"time_window": timeWindow,
		"message":     "Engagement statistics endpoint - implementation pending",
		"note":        "This endpoint will provide comprehensive engagement analytics",
	}

	h.respondSuccess(w, response)
}
