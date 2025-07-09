package repository

import (
	"database/sql"
	"testing"

	"genje-api/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func setupSourceTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE news_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		feed_url TEXT NOT NULL,
		category TEXT DEFAULT 'general',
		active BOOLEAN DEFAULT 1
	);

	CREATE INDEX idx_sources_active ON news_sources(active);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func createTestSource(name, url, feedURL, category string, active bool) models.NewsSource {
	return models.NewsSource{
		Name:     name,
		URL:      url,
		FeedURL:  feedURL,
		Category: category,
		Active:   active,
	}
}

func TestNewSourceRepository(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()

	repo := NewSourceRepository(db)
	if repo == nil {
		t.Error("NewSourceRepository returned nil")
	}
	if repo.db != db {
		t.Error("NewSourceRepository did not set database correctly")
	}
}

func TestCreateSource(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	source := createTestSource("Test Source", "https://example.com", "https://example.com/feed", "news", true)

	err := repo.CreateSource(&source)
	if err != nil {
		t.Fatalf("CreateSource failed: %v", err)
	}

	// Verify source was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM news_sources WHERE name = ?", source.Name).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query source count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 source, got %d", count)
	}
}

func TestGetActiveSources(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	// Create test sources (some active, some inactive)
	sources := []models.NewsSource{
		createTestSource("Active Source 1", "https://example1.com", "https://example1.com/feed", "news", true),
		createTestSource("Inactive Source", "https://example2.com", "https://example2.com/feed", "sports", false),
		createTestSource("Active Source 2", "https://example3.com", "https://example3.com/feed", "tech", true),
	}

	for _, source := range sources {
		err := repo.CreateSource(&source)
		if err != nil {
			t.Fatalf("CreateSource failed: %v", err)
		}
	}

	// Get active sources
	activeSources, err := repo.GetActiveSources()
	if err != nil {
		t.Fatalf("GetActiveSources failed: %v", err)
	}

	// Should only return 2 active sources
	if len(activeSources) != 2 {
		t.Errorf("Expected 2 active sources, got %d", len(activeSources))
	}

	// Verify all returned sources are active
	for _, source := range activeSources {
		if !source.Active {
			t.Errorf("GetActiveSources returned inactive source: %s", source.Name)
		}
	}

	// Verify sources are ordered by name
	if len(activeSources) >= 2 {
		if activeSources[0].Name > activeSources[1].Name {
			t.Error("Sources should be ordered by name")
		}
	}
}

func TestGetSourcesForAggregation(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	// Create test sources
	sources := []models.NewsSource{
		createTestSource("Active Source 1", "https://example1.com", "https://example1.com/feed", "news", true),
		createTestSource("Inactive Source", "https://example2.com", "https://example2.com/feed", "sports", false),
		createTestSource("Active Source 2", "https://example3.com", "https://example3.com/feed", "tech", true),
	}

	for _, source := range sources {
		err := repo.CreateSource(&source)
		if err != nil {
			t.Fatalf("CreateSource failed: %v", err)
		}
	}

	// Get sources for aggregation
	aggSources, err := repo.GetSourcesForAggregation()
	if err != nil {
		t.Fatalf("GetSourcesForAggregation failed: %v", err)
	}

	// Should only return 2 active sources
	if len(aggSources) != 2 {
		t.Errorf("Expected 2 sources for aggregation, got %d", len(aggSources))
	}

	// Verify required fields are populated
	for _, source := range aggSources {
		if source.Name == "" {
			t.Error("Source name should not be empty")
		}
		if source.FeedURL == "" {
			t.Error("Source feed URL should not be empty")
		}
		if source.Category == "" {
			t.Error("Source category should not be empty")
		}
	}
}

func TestSeedInitialSources(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	// Seed initial sources
	err := repo.SeedInitialSources()
	if err != nil {
		t.Fatalf("SeedInitialSources failed: %v", err)
	}

	// Verify sources were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM news_sources").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count sources: %v", err)
	}

	// Should have seeded multiple sources (15 in the current implementation)
	if count < 10 {
		t.Errorf("Expected at least 10 seeded sources, got %d", count)
	}

	// Verify specific sources exist
	expectedSources := []string{
		"The Standard Headlines",
		"Daily Nation",
		"Capital FM News",
		"Business Daily",
	}

	for _, expectedName := range expectedSources {
		var sourceCount int
		err = db.QueryRow("SELECT COUNT(*) FROM news_sources WHERE name = ?", expectedName).Scan(&sourceCount)
		if err != nil {
			t.Fatalf("Failed to check for source %s: %v", expectedName, err)
		}
		if sourceCount != 1 {
			t.Errorf("Expected to find source %s, found %d", expectedName, sourceCount)
		}
	}
}

func TestSeedInitialSourcesIdempotent(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	// Seed initial sources twice
	err := repo.SeedInitialSources()
	if err != nil {
		t.Fatalf("First SeedInitialSources failed: %v", err)
	}

	err = repo.SeedInitialSources()
	if err != nil {
		t.Fatalf("Second SeedInitialSources failed: %v", err)
	}

	// Count sources - should not have duplicates
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM news_sources").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count sources: %v", err)
	}

	// Verify no duplicates were created
	var distinctCount int
	err = db.QueryRow("SELECT COUNT(DISTINCT name) FROM news_sources").Scan(&distinctCount)
	if err != nil {
		t.Fatalf("Failed to count distinct sources: %v", err)
	}

	if count != distinctCount {
		t.Errorf("Duplicate sources created: total=%d, distinct=%d", count, distinctCount)
	}
}

func TestSeedInitialSourcesCategories(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	err := repo.SeedInitialSources()
	if err != nil {
		t.Fatalf("SeedInitialSources failed: %v", err)
	}

	// Check for various categories
	expectedCategories := []string{"news", "sports", "business", "politics", "world"}
	for _, category := range expectedCategories {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM news_sources WHERE category = ?", category).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count sources in category %s: %v", category, err)
		}
		if count == 0 {
			t.Errorf("No sources found in category %s", category)
		}
	}
}

func TestCreateSourceAllFields(t *testing.T) {
	db := setupSourceTestDB(t)
	defer db.Close()
	repo := NewSourceRepository(db)

	source := models.NewsSource{
		Name:     "Detailed Test Source",
		URL:      "https://detailed.example.com",
		FeedURL:  "https://detailed.example.com/rss",
		Category: "technology",
		Active:   false, // Test inactive source
	}

	err := repo.CreateSource(&source)
	if err != nil {
		t.Fatalf("CreateSource failed: %v", err)
	}

	// Verify all fields were saved correctly
	var retrieved models.NewsSource
	err = db.QueryRow(`
		SELECT name, url, feed_url, category, active 
		FROM news_sources WHERE name = ?
	`, source.Name).Scan(&retrieved.Name, &retrieved.URL, &retrieved.FeedURL, &retrieved.Category, &retrieved.Active)

	if err != nil {
		t.Fatalf("Failed to retrieve source: %v", err)
	}

	if retrieved.Name != source.Name {
		t.Errorf("Expected name %s, got %s", source.Name, retrieved.Name)
	}
	if retrieved.URL != source.URL {
		t.Errorf("Expected URL %s, got %s", source.URL, retrieved.URL)
	}
	if retrieved.FeedURL != source.FeedURL {
		t.Errorf("Expected FeedURL %s, got %s", source.FeedURL, retrieved.FeedURL)
	}
	if retrieved.Category != source.Category {
		t.Errorf("Expected category %s, got %s", source.Category, retrieved.Category)
	}
	if retrieved.Active != source.Active {
		t.Errorf("Expected active %v, got %v", source.Active, retrieved.Active)
	}
} 