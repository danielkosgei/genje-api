package repository

import (
	"database/sql"
	"testing"
	"time"

	"genje-api/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT DEFAULT '',
		summary TEXT DEFAULT '',
		url TEXT UNIQUE NOT NULL,
		author TEXT DEFAULT '',
		source TEXT NOT NULL,
		published_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		category TEXT DEFAULT 'general',
		image_url TEXT DEFAULT ''
	);

	CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX idx_articles_category ON articles(category);
	CREATE INDEX idx_articles_source ON articles(source);
	CREATE INDEX idx_articles_url ON articles(url);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func createTestArticle(title, content, url, source, category string) models.Article {
	return models.Article{
		Title:       title,
		Content:     content,
		URL:         url,
		Author:      "Test Author",
		Source:      source,
		PublishedAt: time.Now(),
		Category:    category,
		ImageURL:    "https://example.com/image.jpg",
	}
}

func TestNewArticleRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewArticleRepository(db)
	if repo == nil {
		t.Fatal("NewArticleRepository returned nil")
	}
	if repo.db != db {
		t.Error("NewArticleRepository did not set database correctly")
	}
}

func TestCreateArticle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	article := createTestArticle("Test Title", "Test Content", "https://example.com/test", "TestSource", "news")

	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// Verify article was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE url = ?", article.URL).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query article count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 article, got %d", count)
	}
}

func TestCreateArticleDuplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	article := createTestArticle("Test Title", "Test Content", "https://example.com/test", "TestSource", "news")

	// Create first article
	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// Try to create duplicate (same URL)
	duplicate := article
	duplicate.Title = "Different Title"
	err = repo.CreateArticle(&duplicate)
	if err != nil {
		t.Fatalf("CreateArticle with duplicate URL should not fail (INSERT OR IGNORE): %v", err)
	}

	// Verify only one article exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles WHERE url = ?", article.URL).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query article count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 article after duplicate insert, got %d", count)
	}
}

func TestGetArticleByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	// Create test article
	article := createTestArticle("Test Title", "Test Content", "https://example.com/test", "TestSource", "news")
	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// Get the ID
	var id int
	err = db.QueryRow("SELECT id FROM articles WHERE url = ?", article.URL).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to get article ID: %v", err)
	}

	// Test GetArticleByID
	retrieved, err := repo.GetArticleByID(id)
	if err != nil {
		t.Fatalf("GetArticleByID failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetArticleByID returned nil")
	}

	if retrieved.Title != article.Title {
		t.Errorf("Expected title %s, got %s", article.Title, retrieved.Title)
	}
	if retrieved.URL != article.URL {
		t.Errorf("Expected URL %s, got %s", article.URL, retrieved.URL)
	}
}

func TestGetArticleByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	article, err := repo.GetArticleByID(999)
	if err != nil {
		t.Fatalf("GetArticleByID should not return error for non-existent ID: %v", err)
	}
	if article != nil {
		t.Error("GetArticleByID should return nil for non-existent ID")
	}
}

func TestGetArticlesWithFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	// Create test articles
	articles := []models.Article{
		createTestArticle("Sports News 1", "Basketball content", "https://example.com/sports1", "SportsSource", "sports"),
		createTestArticle("Tech News 1", "Technology content", "https://example.com/tech1", "TechSource", "technology"),
		createTestArticle("Sports News 2", "Football content", "https://example.com/sports2", "SportsSource", "sports"),
		createTestArticle("Politics News", "Government content", "https://example.com/politics1", "NewsSource", "politics"),
	}

	for _, article := range articles {
		err := repo.CreateArticle(&article)
		if err != nil {
			t.Fatalf("CreateArticle failed: %v", err)
		}
	}

	tests := []struct {
		name     string
		filters  models.ArticleFilters
		expected int
	}{
		{
			name:     "No filters",
			filters:  models.ArticleFilters{Page: 1, Limit: 10},
			expected: 4,
		},
		{
			name:     "Filter by category",
			filters:  models.ArticleFilters{Page: 1, Limit: 10, Category: "sports"},
			expected: 2,
		},
		{
			name:     "Filter by source",
			filters:  models.ArticleFilters{Page: 1, Limit: 10, Source: "SportsSource"},
			expected: 2,
		},
		{
			name:     "Search in content",
			filters:  models.ArticleFilters{Page: 1, Limit: 10, Search: "Basketball"},
			expected: 1,
		},
		{
			name:     "Pagination",
			filters:  models.ArticleFilters{Page: 1, Limit: 2},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles, total, err := repo.GetArticles(tt.filters)
			if err != nil {
				t.Fatalf("GetArticles failed: %v", err)
			}

			if len(articles) != tt.expected {
				t.Errorf("Expected %d articles, got %d", tt.expected, len(articles))
			}

			if tt.name == "No filters" && total != 4 {
				t.Errorf("Expected total count 4, got %d", total)
			}
		})
	}
}

func TestUpdateSummary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	// Create test article
	article := createTestArticle("Test Title", "Test Content", "https://example.com/test", "TestSource", "news")
	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// Get the ID
	var id int
	err = db.QueryRow("SELECT id FROM articles WHERE url = ?", article.URL).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to get article ID: %v", err)
	}

	// Update summary
	summary := "This is a test summary"
	err = repo.UpdateSummary(id, summary)
	if err != nil {
		t.Fatalf("UpdateSummary failed: %v", err)
	}

	// Verify summary was updated
	var retrievedSummary string
	err = db.QueryRow("SELECT summary FROM articles WHERE id = ?", id).Scan(&retrievedSummary)
	if err != nil {
		t.Fatalf("Failed to retrieve summary: %v", err)
	}
	if retrievedSummary != summary {
		t.Errorf("Expected summary %s, got %s", summary, retrievedSummary)
	}
}

func TestGetCategories(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	// Create articles with different categories
	articles := []models.Article{
		createTestArticle("Sports News", "Content", "https://example.com/sports", "Source", "sports"),
		createTestArticle("Tech News", "Content", "https://example.com/tech", "Source", "technology"),
		createTestArticle("Sports News 2", "Content", "https://example.com/sports2", "Source", "sports"),
	}

	for _, article := range articles {
		err := repo.CreateArticle(&article)
		if err != nil {
			t.Fatalf("CreateArticle failed: %v", err)
		}
	}

	categories, err := repo.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	expectedCategories := []string{"sports", "technology"}
	if len(categories) != len(expectedCategories) {
		t.Errorf("Expected %d categories, got %d", len(expectedCategories), len(categories))
	}

	// Check if all expected categories are present
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}

	for _, expected := range expectedCategories {
		if !categoryMap[expected] {
			t.Errorf("Expected category %s not found", expected)
		}
	}
}

func TestCreateArticlesBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	articles := []models.Article{
		createTestArticle("Article 1", "Content 1", "https://example.com/1", "Source1", "news"),
		createTestArticle("Article 2", "Content 2", "https://example.com/2", "Source2", "sports"),
		createTestArticle("Article 3", "Content 3", "https://example.com/3", "Source3", "tech"),
	}

	err := repo.CreateArticlesBatch(articles)
	if err != nil {
		t.Fatalf("CreateArticlesBatch failed: %v", err)
	}

	// Verify all articles were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 articles, got %d", count)
	}
}

func TestCreateArticlesBatchEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	err := repo.CreateArticlesBatch([]models.Article{})
	if err != nil {
		t.Fatalf("CreateArticlesBatch with empty slice should not fail: %v", err)
	}
}

func TestCreateArticlesBatchWithDuplicates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewArticleRepository(db)

	// Create first article
	article1 := createTestArticle("Article 1", "Content 1", "https://example.com/1", "Source1", "news")
	err := repo.CreateArticle(&article1)
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	// Try to batch create with duplicate and new article
	articles := []models.Article{
		article1, // duplicate
		createTestArticle("Article 2", "Content 2", "https://example.com/2", "Source2", "sports"),
	}

	err = repo.CreateArticlesBatch(articles)
	if err != nil {
		t.Fatalf("CreateArticlesBatch should handle duplicates gracefully: %v", err)
	}

	// Should have 2 articles total (duplicate ignored)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count articles: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 articles (duplicate ignored), got %d", count)
	}
} 