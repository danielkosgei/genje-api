package database

import (
	"testing"
	"time"

	"genje-api/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	// Test that the database connection is working
	err = db.Ping()
	if err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Verify connection pool settings
	stats := db.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections 10, got %d", stats.MaxOpenConnections)
	}
}

func TestNewWithInvalidURL(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             "/invalid/path/that/does/not/exist/test.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err == nil {
		if db != nil {
			db.Close()
		}
		// Note: SQLite might not fail immediately on invalid paths,
		// so we'll test with ping instead
		if db != nil {
			err = db.Ping()
			if err == nil {
				t.Error("Expected error for invalid database path")
			}
		}
	}
	// If we get an error during New(), that's also acceptable
}

func TestMigrate(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify that tables were created
	tables := []string{"articles", "news_sources"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s should exist after migration", table)
		}
	}
}

func TestMigrateArticlesTableStructure(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Check articles table columns
	expectedColumns := []string{
		"id", "title", "content", "summary", "url", "author",
		"source", "published_at", "created_at", "category", "image_url",
	}

	rows, err := db.Query("PRAGMA table_info(articles)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	var foundColumns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		foundColumns = append(foundColumns, name)
	}

	for _, expectedCol := range expectedColumns {
		found := false
		for _, foundCol := range foundColumns {
			if foundCol == expectedCol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column %s not found in articles table", expectedCol)
		}
	}
}

func TestMigrateNewsSourcesTableStructure(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Check news_sources table columns
	expectedColumns := []string{"id", "name", "url", "feed_url", "category", "active"}

	rows, err := db.Query("PRAGMA table_info(news_sources)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	var foundColumns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		foundColumns = append(foundColumns, name)
	}

	for _, expectedCol := range expectedColumns {
		found := false
		for _, foundCol := range foundColumns {
			if foundCol == expectedCol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column %s not found in news_sources table", expectedCol)
		}
	}
}

func TestMigrateIndexes(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Check that indexes were created
	expectedIndexes := []string{
		"idx_articles_published_at",
		"idx_articles_category",
		"idx_articles_source",
		"idx_articles_url",
		"idx_sources_active",
	}

	for _, indexName := range expectedIndexes {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for index %s: %v", indexName, err)
		}
		if count != 1 {
			t.Errorf("Index %s should exist after migration", indexName)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	// Run migrations twice
	err = Migrate(db)
	if err != nil {
		t.Fatalf("First Migrate() failed: %v", err)
	}

	err = Migrate(db)
	if err != nil {
		t.Fatalf("Second Migrate() should not fail: %v", err)
	}

	// Verify tables still exist and are not duplicated
	var articlesCount, sourcesCount int

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='articles'").Scan(&articlesCount)
	if err != nil {
		t.Fatalf("Failed to count articles table: %v", err)
	}
	if articlesCount != 1 {
		t.Errorf("Should have exactly 1 articles table, got %d", articlesCount)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='news_sources'").Scan(&sourcesCount)
	if err != nil {
		t.Fatalf("Failed to count news_sources table: %v", err)
	}
	if sourcesCount != 1 {
		t.Errorf("Should have exactly 1 news_sources table, got %d", sourcesCount)
	}
}

func TestConnectionLimits(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL:             ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    3,
		ConnMaxLifetime: 1 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	// Check that connection limits are properly set
	stats := db.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("Expected MaxOpenConnections 5, got %d", stats.MaxOpenConnections)
	}

	// Test that we can actually use the database
	err = db.Ping()
	if err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}
}

func TestDatabaseFile(t *testing.T) {
	// Test with an actual file database (not in-memory)
	cfg := config.DatabaseConfig{
		URL:             "./test.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	err = Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Clear any existing data first
	_, err = db.Exec("DELETE FROM articles")
	if err != nil {
		t.Fatalf("Failed to clear articles: %v", err)
	}

	// Test basic operations
	_, err = db.Exec("INSERT INTO articles (title, url, source) VALUES (?, ?, ?)", "Test File DB", "http://test-file.com", "Test Source")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 article, got %d", count)
	}

	db.Close()

	// Clean up test file
	// Note: In a real test, you might want to use a temporary directory
}
