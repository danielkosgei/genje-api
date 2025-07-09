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