package models

// ArticleFiltersRequest represents query parameters for article filtering
type ArticleFiltersRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	Limit    int    `json:"limit" validate:"min=1,max=100"`
	Category string `json:"category"`
	Source   string `json:"source"`
	Search   string `json:"search"`
	From     string `json:"from" validate:"omitempty,datetime=2006-01-02"`
	To       string `json:"to" validate:"omitempty,datetime=2006-01-02"`
	SortBy   string `json:"sort_by" validate:"omitempty,oneof=date title source"`
	SortDir  string `json:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

// SearchRequest represents search parameters
type SearchRequest struct {
	Query     string `json:"q" validate:"required,min=2"`
	Category  string `json:"category"`
	Source    string `json:"source"`
	From      string `json:"from" validate:"omitempty,datetime=2006-01-02"`
	To        string `json:"to" validate:"omitempty,datetime=2006-01-02"`
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=relevance date source"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// FeedRequestParams represents feed parameters for validation
type FeedRequestParams struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Category  string `json:"category"`
	Source    string `json:"source"`
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=date popularity"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// TrendingRequest represents trending parameters
type TrendingRequest struct {
	Limit      int    `json:"limit" validate:"min=1,max=100"`
	TimeWindow string `json:"window" validate:"omitempty,oneof=1h 6h 12h 24h 7d"`
}

// StatsRequest represents statistics parameters
type StatsRequest struct {
	Days int `json:"days" validate:"min=1,max=365"`
}

// Default values
const (
	DefaultPage      = 1
	DefaultLimit     = 20
	MaxLimit         = 100
	DefaultSortBy    = "date"
	DefaultSortOrder = "desc"
	DefaultHours     = 24
	DefaultDays      = 30
	DefaultWindow    = "24h"
)
