package database

import (
	"database/sql"
	"fmt"

	"genje-api/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

func New(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection limits
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		summary TEXT,
		url TEXT UNIQUE NOT NULL,
		author TEXT,
		source TEXT NOT NULL,
		published_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		category TEXT DEFAULT 'general',
		image_url TEXT
	);

	CREATE TABLE IF NOT EXISTS news_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		feed_url TEXT NOT NULL,
		category TEXT DEFAULT 'general',
		active BOOLEAN DEFAULT 1
	);

	-- Engagement tracking tables
	CREATE TABLE IF NOT EXISTS article_engagement (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		views INTEGER DEFAULT 0,
		shares INTEGER DEFAULT 0,
		comments INTEGER DEFAULT 0,
		likes INTEGER DEFAULT 0,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS engagement_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		event_type TEXT NOT NULL, -- 'view', 'share', 'comment', 'like'
		user_ip TEXT, -- For basic tracking without user accounts
		user_agent TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT, -- JSON for additional data
		FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	-- Source authority scoring
	CREATE TABLE IF NOT EXISTS source_authority (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT UNIQUE NOT NULL,
		authority_score REAL DEFAULT 0.5, -- 0.0 to 1.0
		credibility_score REAL DEFAULT 0.5,
		reach_score REAL DEFAULT 0.5,
		total_articles INTEGER DEFAULT 0,
		avg_engagement REAL DEFAULT 0.0,
		last_calculated DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Trending cache for performance
	CREATE TABLE IF NOT EXISTS trending_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		time_window TEXT NOT NULL, -- '1h', '6h', '24h', '7d'
		trending_score REAL NOT NULL,
		engagement_velocity REAL DEFAULT 0.0,
		recency_score REAL DEFAULT 0.0,
		authority_score REAL DEFAULT 0.0,
		content_score REAL DEFAULT 0.0,
		calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
	CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source);
	CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
	CREATE INDEX IF NOT EXISTS idx_sources_active ON news_sources(active);
	
	-- Engagement indexes
	CREATE UNIQUE INDEX IF NOT EXISTS idx_engagement_article ON article_engagement(article_id);
	CREATE INDEX IF NOT EXISTS idx_engagement_events_article ON engagement_events(article_id);
	CREATE INDEX IF NOT EXISTS idx_engagement_events_type ON engagement_events(event_type);
	CREATE INDEX IF NOT EXISTS idx_engagement_events_timestamp ON engagement_events(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_source_authority_name ON source_authority(source_name);
	CREATE INDEX IF NOT EXISTS idx_trending_cache_window ON trending_cache(time_window, trending_score DESC);
	CREATE INDEX IF NOT EXISTS idx_trending_cache_calculated ON trending_cache(calculated_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
