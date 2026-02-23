package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SocialPost struct {
	ID             uuid.UUID       `json:"id"`
	Platform       string          `json:"platform"`
	PlatformPostID *string         `json:"platform_post_id,omitempty"`
	AuthorHandle   *string         `json:"author_handle,omitempty"`
	AuthorName     *string         `json:"author_name,omitempty"`
	Content        string          `json:"content"`
	URL            *string         `json:"url,omitempty"`
	PostedAt       *time.Time      `json:"posted_at,omitempty"`
	ScrapedAt      time.Time       `json:"scraped_at"`
	Engagement     json.RawMessage `json:"engagement"`
	CreatedAt      time.Time       `json:"created_at"`
}

type SocialPostMention struct {
	PostID         uuid.UUID `json:"post_id"`
	PoliticianID   uuid.UUID `json:"politician_id"`
	SentimentScore *float64  `json:"sentiment_score,omitempty"`
}
