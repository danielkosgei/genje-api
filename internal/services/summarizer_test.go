package services

import (
	"database/sql"
	"testing"
	"time"

	"genje-api/internal/models"
	"genje-api/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func setupSummarizerTestDB(t *testing.T) (*sql.DB, *repository.ArticleRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create articles table
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
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db, repository.NewArticleRepository(db)
}

func createTestArticleForSummarizer(t *testing.T, db *sql.DB, repo *repository.ArticleRepository, title, content, url string) int {
	article := models.Article{
		Title:       title,
		Content:     content,
		URL:         url,
		Author:      "Test Author",
		Source:      "Test Source",
		PublishedAt: time.Now(),
		Category:    "test",
	}

	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	// Get the created article ID
	createdArticle, err := repo.GetArticleByID(1) // Assuming first article gets ID 1
	if err != nil {
		t.Fatalf("Failed to get created article: %v", err)
	}
	if createdArticle == nil {
		// Try to get the ID directly from database using raw SQL
		var id int
		err = db.QueryRow("SELECT id FROM articles WHERE url = ?", url).Scan(&id)
		if err != nil {
			t.Fatalf("Failed to get article ID: %v", err)
		}
		return id
	}

	return createdArticle.ID
}

func TestNewSummarizerService(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()

	service := NewSummarizerService(repo)
	if service == nil {
		t.Error("NewSummarizerService returned nil")
	}
	if service.articleRepo != repo {
		t.Error("NewSummarizerService did not set repository correctly")
	}
}

func TestSummarizeArticle(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	content := "This is the first sentence of the article. This is the second sentence with more details. This is the third sentence continuing the story. This is the fourth sentence that should not be included in the summary."
	articleID := createTestArticleForSummarizer(t, db, repo, "Test Article", content, "https://example.com/test")

	summary, err := service.SummarizeArticle(articleID)
	if err != nil {
		t.Fatalf("SummarizeArticle failed: %v", err)
	}

	if summary == "" {
		t.Error("Summary should not be empty")
	}

	// Should contain the first three sentences
	expectedParts := []string{
		"This is the first sentence of the article",
		"This is the second sentence with more details",
		"This is the third sentence continuing the story",
	}

	for _, part := range expectedParts {
		if !contains(summary, part) {
			t.Errorf("Summary should contain: %s", part)
		}
	}

	// Should not contain the fourth sentence
	if contains(summary, "This is the fourth sentence") {
		t.Error("Summary should not contain the fourth sentence")
	}
}

func TestSummarizeArticleWithExistingSummary(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	content := "This is test content for the article."
	articleID := createTestArticleForSummarizer(t, db, repo, "Test Article", content, "https://example.com/test")

	// Create a summary first
	existingSummary := "This is an existing summary."
	err := repo.UpdateSummary(articleID, existingSummary)
	if err != nil {
		t.Fatalf("Failed to set existing summary: %v", err)
	}

	// Request summary again - should return existing one
	summary, err := service.SummarizeArticle(articleID)
	if err != nil {
		t.Fatalf("SummarizeArticle failed: %v", err)
	}

	if summary != existingSummary {
		t.Errorf("Expected existing summary %s, got %s", existingSummary, summary)
	}
}

func TestSummarizeArticleNotFound(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	summary, err := service.SummarizeArticle(999)
	if err == nil {
		t.Error("SummarizeArticle should return error for non-existent article")
	}
	if summary != "" {
		t.Error("Summary should be empty for non-existent article")
	}
	if err.Error() != "article not found" {
		t.Errorf("Expected 'article not found' error, got: %v", err)
	}
}

func TestGenerateSimpleSummary(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "Single sentence",
			content:  "This is a single sentence.",
			expected: "This is a single sentence.",
		},
		{
			name:     "Two sentences",
			content:  "First sentence. Second sentence.",
			expected: "First sentence. Second sentence.",
		},
		{
			name:     "Three sentences",
			content:  "First sentence. Second sentence. Third sentence.",
			expected: "First sentence. Second sentence. Third sentence.",
		},
		{
			name:     "Four sentences - should truncate",
			content:  "First sentence. Second sentence. Third sentence. Fourth sentence.",
			expected: "First sentence. Second sentence. Third sentence.",
		},
		{
			name:     "Content with HTML-like tags",
			content:  "This has <tags> in it. Another sentence with <more>tags</more>. Third sentence normal.",
			expected: "This has  <tags>  in it. Another sentence with  <more> tags </more> . Third sentence normal.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateSimpleSummary(tt.content)
			if result != tt.expected {
				t.Errorf("Expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateSimpleSummaryLongContent(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	// Create content that would exceed 300 characters
	longSentence := "This is a very long sentence that contains a lot of words and goes on and on and on to test the character limit functionality of the summarizer service which should truncate content that exceeds the maximum allowed length of three hundred characters total."
	content := longSentence + " Second sentence. Third sentence."

	summary := service.generateSimpleSummary(content)

	// Should be limited to 300 characters
	if len(summary) > 300 {
		t.Errorf("Summary should be limited to 300 characters, got %d", len(summary))
	}

	// Should end with "..."
	if len(summary) == 300 && !endsWith(summary, "...") {
		t.Error("Truncated summary should end with '...'")
	}
}

func TestGenerateSimpleSummaryWhitespace(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Leading/trailing whitespace",
			content:  "   First sentence. Second sentence.   ",
			expected: "First sentence. Second sentence.   .",
		},
		{
			name:     "Multiple spaces",
			content:  "First    sentence.    Second sentence.",
			expected: "First    sentence.    Second sentence.",
		},
		{
			name:     "Newlines and tabs",
			content:  "First sentence.\n\tSecond sentence.",
			expected: "First sentence.\n\tSecond sentence.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateSimpleSummary(tt.content)
			if result != tt.expected {
				t.Errorf("Expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}

func TestSummarizeArticleEmptyContent(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	articleID := createTestArticleForSummarizer(t, db, repo, "Test Article", "", "https://example.com/empty")

	summary, err := service.SummarizeArticle(articleID)
	if err != nil {
		t.Fatalf("SummarizeArticle should not fail for empty content: %v", err)
	}

	if summary != "" {
		t.Errorf("Summary should be empty for empty content, got: %s", summary)
	}
}

func TestSummarizeArticleUpdatesDatabase(t *testing.T) {
	db, repo := setupSummarizerTestDB(t)
	defer db.Close()
	service := NewSummarizerService(repo)

	content := "First sentence. Second sentence. Third sentence."
	articleID := createTestArticleForSummarizer(t, db, repo, "Test Article", content, "https://example.com/test")

	summary, err := service.SummarizeArticle(articleID)
	if err != nil {
		t.Fatalf("SummarizeArticle failed: %v", err)
	}

	// Verify the summary was saved to the database
	article, err := repo.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("Failed to retrieve article: %v", err)
	}
	if article == nil {
		t.Fatal("Article should not be nil")
	}

	if article.Summary != summary {
		t.Errorf("Article summary in database %s does not match returned summary %s", article.Summary, summary)
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
} 