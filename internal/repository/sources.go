package repository

import (
	"database/sql"
	"fmt"

	"genje-api/internal/models"
)

type SourceRepository struct {
	db *sql.DB
}

func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) GetActiveSources() ([]models.NewsSource, error) {
	query := "SELECT id, name, url, feed_url, category, active FROM news_sources WHERE active = 1 ORDER BY name"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var source models.NewsSource
		err := rows.Scan(&source.ID, &source.Name, &source.URL, &source.FeedURL, 
			&source.Category, &source.Active)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return sources, nil
}

func (r *SourceRepository) GetSourcesForAggregation() ([]models.NewsSource, error) {
	query := "SELECT name, feed_url, category FROM news_sources WHERE active = 1"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources for aggregation: %w", err)
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var source models.NewsSource
		err := rows.Scan(&source.Name, &source.FeedURL, &source.Category)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

func (r *SourceRepository) CreateSource(source *models.NewsSource) error {
	query := `
		INSERT INTO news_sources (name, url, feed_url, category, active)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, source.Name, source.URL, source.FeedURL, source.Category, source.Active)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	return nil
}

func (r *SourceRepository) SeedInitialSources() error {
	sources := []models.NewsSource{
		{Name: "The Standard Headlines", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/headlines.php", Category: "news", Active: true},
		{Name: "Standard Kenya", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/kenya.php", Category: "news", Active: true},
		{Name: "Standard World", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/world.php", Category: "world", Active: true},
		{Name: "Standard Politics", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/politics.php", Category: "politics", Active: true},
		{Name: "Standard Sports", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/sports.php", Category: "sports", Active: true},
		{Name: "Standard Business", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/business.php", Category: "business", Active: true},
		{Name: "Standard Columnists", URL: "https://standardmedia.co.ke", FeedURL: "https://www.standardmedia.co.ke/rss/columnists.php", Category: "opinion", Active: true},
		{Name: "Daily Nation", URL: "https://nation.africa", FeedURL: "https://nation.africa/kenya/rss.xml", Category: "news", Active: true},
		{Name: "Taifa Leo", URL: "https://taifaleo.nation.co.ke", FeedURL: "https://taifaleo.nation.co.ke/feed", Category: "kiswahili", Active: true},
		{Name: "Capital FM News", URL: "https://capitalfm.co.ke", FeedURL: "https://capitalfm.co.ke/news/feed", Category: "news", Active: true},
		{Name: "Nairobi Wire", URL: "https://nairobiwire.com", FeedURL: "https://nairobiwire.com/feed", Category: "news", Active: true},
		{Name: "Diaspora Messenger", URL: "https://diasporamessenger.com", FeedURL: "https://diasporamessenger.com/feed", Category: "diaspora", Active: true},
		{Name: "Sharp Daily", URL: "https://thesharpdaily.com", FeedURL: "https://thesharpdaily.com/feed", Category: "news", Active: true},
		{Name: "Kenyans.co.ke", URL: "https://kenyans.co.ke", FeedURL: "https://www.kenyans.co.ke/feeds/news", Category: "news", Active: true},
		{Name: "Business Daily", URL: "https://www.businessdailyafrica.com", FeedURL: "https://www.businessdailyafrica.com/rss", Category: "business", Active: true},
	}

	for _, src := range sources {
		// Check if source already exists by name and feed_url
		var count int
		err := r.db.QueryRow(
			`SELECT COUNT(*) FROM news_sources WHERE name = ? AND feed_url = ?`,
			src.Name, src.FeedURL,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check existing source %s: %w", src.Name, err)
		}
		
		// Only insert if it doesn't exist
		if count == 0 {
			_, err := r.db.Exec(
				`INSERT INTO news_sources (name, url, feed_url, category, active)
				 VALUES (?, ?, ?, ?, ?)`,
				src.Name, src.URL, src.FeedURL, src.Category, src.Active,
			)
			if err != nil {
				return fmt.Errorf("failed to seed source %s: %w", src.Name, err)
			}
		}
	}

	return nil
} 