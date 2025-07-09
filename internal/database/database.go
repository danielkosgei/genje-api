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

	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
	CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source);
	CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
	CREATE INDEX IF NOT EXISTS idx_sources_active ON news_sources(active);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
} 