package models

import (
	"time"
)

type Article struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	Summary     string    `json:"summary" db:"summary"`
	URL         string    `json:"url" db:"url"`
	Author      string    `json:"author" db:"author"`
	Source      string    `json:"source" db:"source"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Category    string    `json:"category" db:"category"`
	ImageURL    string    `json:"image_url" db:"image_url"`
}

type NewsSource struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	URL      string `json:"url" db:"url"`
	FeedURL  string `json:"feed_url" db:"feed_url"`
	Category string `json:"category" db:"category"`
	Active   bool   `json:"active" db:"active"`
}

type ArticleFilters struct {
	Page     int
	Limit    int
	Category string
	Source   string
	Search   string
	From     string // Date range filtering
	To       string
}

type PaginationResponse struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ArticlesResponse struct {
	Articles   []Article          `json:"articles"`
	Pagination PaginationResponse `json:"pagination"`
}

type SourcesResponse struct {
	Sources []NewsSource `json:"sources"`
}

type CategoriesResponse struct {
	Categories []string `json:"categories"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// New enhanced models
type APIInfoResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Endpoints   []string  `json:"endpoints"`
	LastUpdated time.Time `json:"last_updated"`
}

type GlobalStats struct {
	TotalArticles int `json:"total_articles"`
	TotalSources  int `json:"total_sources"`
	Categories    int `json:"categories"`
	LastUpdated   time.Time `json:"last_updated"`
}

type SourceStats struct {
	Name         string `json:"name"`
	ArticleCount int    `json:"article_count"`
	Category     string `json:"category"`
	LastUpdated  time.Time `json:"last_updated"`
}

type CategoryStats struct {
	Category     string `json:"category"`
	ArticleCount int    `json:"article_count"`
}

type TimelineStats struct {
	Date         string `json:"date"`
	ArticleCount int    `json:"article_count"`
}

type SystemStatus struct {
	Status           string    `json:"status"`
	LastAggregation  time.Time `json:"last_aggregation"`
	ActiveSources    int       `json:"active_sources"`
	TotalArticles    int       `json:"total_articles"`
	AggregationError string    `json:"aggregation_error,omitempty"`
}

type StatsResponse struct {
	Success bool        `json:"success"`
	Data    GlobalStats `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
	} `json:"meta"`
}

type SourceStatsResponse struct {
	Success bool          `json:"success"`
	Data    []SourceStats `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
	} `json:"meta"`
}

type CategoryStatsResponse struct {
	Success bool            `json:"success"`
	Data    []CategoryStats `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
	} `json:"meta"`
}

type TimelineStatsResponse struct {
	Success bool            `json:"success"`
	Data    []TimelineStats `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
		Days        int       `json:"days"`
	} `json:"meta"`
}

type RecentArticlesResponse struct {
	Success bool      `json:"success"`
	Data    []Article `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
		Hours       int       `json:"hours"`
		Total       int       `json:"total"`
	} `json:"meta"`
}

type EnhancedArticlesResponse struct {
	Success bool      `json:"success"`
	Data    []Article `json:"data"`
	Meta    struct {
		Pagination  PaginationResponse `json:"pagination"`
		GeneratedAt time.Time          `json:"generated_at"`
		Filters     ArticleFilters     `json:"filters"`
	} `json:"meta"`
}

// Create/Update source request
type CreateSourceRequest struct {
	Name     string `json:"name" validate:"required"`
	URL      string `json:"url" validate:"required,url"`
	FeedURL  string `json:"feed_url" validate:"required,url"`
	Category string `json:"category" validate:"required"`
	Active   bool   `json:"active"`
}

type UpdateSourceRequest struct {
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`
	FeedURL  string `json:"feed_url,omitempty"`
	Category string `json:"category,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// New models for missing endpoints
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    OpenAPIInfo           `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Contact     OpenAPIContact `json:"contact"`
}

type OpenAPIContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

type APISchema struct {
	Models map[string]interface{} `json:"models"`
	Endpoints []EndpointSchema `json:"endpoints"`
}

type EndpointSchema struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Description string                 `json:"description"`
	Parameters  []ParameterSchema      `json:"parameters,omitempty"`
	Response    map[string]interface{} `json:"response"`
}

type ParameterSchema struct {
	Name        string `json:"name"`
	In          string `json:"in"` // "query", "path", "header"
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CursorPaginationResponse struct {
	Articles   []Article `json:"articles"`
	NextCursor string    `json:"next_cursor,omitempty"`
	PrevCursor string    `json:"prev_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
	Total      int       `json:"total"`
}

type TrendingArticle struct {
	Article
	Score           float64 `json:"score"`
	ViewCount       int     `json:"view_count"`
	ShareCount      int     `json:"share_count"`
	CommentCount    int     `json:"comment_count"`
	TrendingReason  string  `json:"trending_reason"`
}

type TrendingTopic struct {
	Topic       string  `json:"topic"`
	Count       int     `json:"count"`
	Score       float64 `json:"score"`
	Articles    []int   `json:"articles"` // Article IDs
	Category    string  `json:"category"`
	FirstSeen   time.Time `json:"first_seen"`
	LastUpdated time.Time `json:"last_updated"`
}

type SearchFilters struct {
	Query       string `json:"query"`
	Category    string `json:"category,omitempty"`
	Source      string `json:"source,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	SortBy      string `json:"sort_by"` // "relevance", "date", "source"
	SortOrder   string `json:"sort_order"` // "asc", "desc"
}

type FeedRequest struct {
	Cursor    string `json:"cursor,omitempty"`
	Limit     int    `json:"limit"`
	Category  string `json:"category,omitempty"`
	Source    string `json:"source,omitempty"`
	SortBy    string `json:"sort_by"` // "date", "popularity"
	SortOrder string `json:"sort_order"` // "asc", "desc"
}

// Response wrapper types
type SearchResponse struct {
	Success bool      `json:"success"`
	Data    []Article `json:"data"`
	Meta    struct {
		Pagination  PaginationResponse `json:"pagination"`
		GeneratedAt time.Time          `json:"generated_at"`
		Query       string             `json:"query"`
		Filters     SearchFilters      `json:"filters"`
	} `json:"meta"`
}

type FeedResponse struct {
	Success bool                     `json:"success"`
	Data    CursorPaginationResponse `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
		Filters     FeedRequest `json:"filters"`
	} `json:"meta"`
}

type TrendingResponse struct {
	Success bool              `json:"success"`
	Data    []TrendingArticle `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
		Algorithm   string    `json:"algorithm"`
		TimeWindow  string    `json:"time_window"`
	} `json:"meta"`
}

type TrendsResponse struct {
	Success bool            `json:"success"`
	Data    []TrendingTopic `json:"data"`
	Meta    struct {
		GeneratedAt time.Time `json:"generated_at"`
		TimeWindow  string    `json:"time_window"`
		Algorithm   string    `json:"algorithm"`
	} `json:"meta"`
} 