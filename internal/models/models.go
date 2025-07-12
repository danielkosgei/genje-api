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