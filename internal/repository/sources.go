package repository

import (
	"database/sql"
	"fmt"
	"strings"

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

// GetSourceByID returns a source by its ID
func (r *SourceRepository) GetSourceByID(id int) (*models.NewsSource, error) {
	query := "SELECT id, name, url, feed_url, category, active FROM news_sources WHERE id = ?"
	
	var source models.NewsSource
	err := r.db.QueryRow(query, id).Scan(&source.ID, &source.Name, &source.URL, 
		&source.FeedURL, &source.Category, &source.Active)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	
	return &source, nil
}

// UpdateSource updates an existing source
func (r *SourceRepository) UpdateSource(id int, updates models.UpdateSourceRequest) error {
	// Use a secure approach that avoids dynamic SQL construction
	// We'll update all fields at once using COALESCE to keep existing values
	
	// First, get the current source to use as defaults
	current, err := r.GetSourceByID(id)
	if err != nil {
		return fmt.Errorf("failed to get current source: %w", err)
	}
	if current == nil {
		return fmt.Errorf("source not found")
	}
	
	// Use provided values or fall back to current values
	name := current.Name
	if updates.Name != "" {
		name = updates.Name
	}
	
	url := current.URL
	if updates.URL != "" {
		url = updates.URL
	}
	
	feedURL := current.FeedURL
	if updates.FeedURL != "" {
		feedURL = updates.FeedURL
	}
	
	category := current.Category
	if updates.Category != "" {
		category = updates.Category
	}
	
	active := current.Active
	if updates.Active != nil {
		active = *updates.Active
	}
	
	// Check if any field actually changed
	if name == current.Name && url == current.URL && feedURL == current.FeedURL && 
		category == current.Category && active == current.Active {
		return fmt.Errorf("no fields to update")
	}
	
	// Static query - no dynamic SQL construction
	query := `UPDATE news_sources SET 
		name = ?, 
		url = ?, 
		feed_url = ?, 
		category = ?, 
		active = ? 
		WHERE id = ?`
	
	result, err := r.db.Exec(query, name, url, feedURL, category, active, id)
	if err != nil {
		return fmt.Errorf("failed to update source: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("source not found")
	}
	
	return nil
}

// DeleteSource soft deletes a source (sets active = false)
func (r *SourceRepository) DeleteSource(id int) error {
	query := "UPDATE news_sources SET active = 0 WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("source not found")
	}
	
	return nil
}

// GetAllSources returns all sources (active and inactive)
func (r *SourceRepository) GetAllSources() ([]models.NewsSource, error) {
	query := "SELECT id, name, url, feed_url, category, active FROM news_sources ORDER BY name"
	
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

// GetSourcesCount returns the total number of sources
func (r *SourceRepository) GetSourcesCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM news_sources WHERE active = 1").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get sources count: %w", err)
	}
	return count, nil
}

// RefreshSingleSource triggers aggregation for a single source
func (r *SourceRepository) RefreshSingleSource(id int) (*models.NewsSource, error) {
	source, err := r.GetSourceByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	
	if source == nil {
		return nil, fmt.Errorf("source not found")
	}
	
	if !source.Active {
		return nil, fmt.Errorf("source is not active")
	}
	
	return source, nil
} 