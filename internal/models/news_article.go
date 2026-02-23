package models

import (
	"time"

	"github.com/google/uuid"
)

type NewsSource struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	FeedURL   *string   `json:"feed_url,omitempty"`
	Type      string    `json:"type"`
	Outlet    *string   `json:"outlet,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NewsArticle struct {
	ID                uuid.UUID  `json:"id"`
	SourceID          *uuid.UUID `json:"source_id,omitempty"`
	Title             string     `json:"title"`
	Content           *string    `json:"content,omitempty"`
	Summary           *string    `json:"summary,omitempty"`
	URL               string     `json:"url"`
	Author            *string    `json:"author,omitempty"`
	ImageURL          *string    `json:"image_url,omitempty"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	ScrapedAt         time.Time  `json:"scraped_at"`
	Category          *string    `json:"category,omitempty"`
	IsElectionRelated bool       `json:"is_election_related"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ArticlePoliticianMention struct {
	ArticleID      uuid.UUID `json:"article_id"`
	PoliticianID   uuid.UUID `json:"politician_id"`
	SentimentScore *float64  `json:"sentiment_score,omitempty"`
}

type NewsFilter struct {
	PoliticianID    *uuid.UUID
	SourceID        *uuid.UUID
	ElectionRelated *bool
	Category        *string
	Since           *time.Time
	Until           *time.Time
	Limit           int
	Offset          int
}
