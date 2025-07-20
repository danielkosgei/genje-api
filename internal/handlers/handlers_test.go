package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"genje-api/internal/config"
	"genje-api/internal/models"
	"genje-api/internal/repository"
	"genje-api/internal/services"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

func setupHandlerTest(t *testing.T) (*sql.DB, *Handler) {
	// Create in-memory database
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

	CREATE TABLE news_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		feed_url TEXT NOT NULL,
		category TEXT DEFAULT 'general',
		active BOOLEAN DEFAULT 1
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Create repositories
	articleRepo := repository.NewArticleRepository(db)
	sourceRepo := repository.NewSourceRepository(db)

	// Create services
	cfg := config.AggregatorConfig{
		Interval:       30 * time.Minute,
		RequestTimeout: 30 * time.Second,
		UserAgent:      "test-agent",
		MaxContentSize: 10000,
		MaxSummarySize: 300,
	}
	aggregatorService := services.NewAggregatorService(articleRepo, sourceRepo, cfg)
	summarizerService := services.NewSummarizerService(articleRepo)
	engagementRepo := repository.NewEngagementRepository(db)
	trendingService := services.NewTrendingService(db, articleRepo, engagementRepo, summarizerService)

	// Create handler
	handler := New(articleRepo, sourceRepo, engagementRepo, aggregatorService, summarizerService, trendingService)

	return db, handler
}

func createTestArticleForHandler(t *testing.T, repo *repository.ArticleRepository, title, content string) int {
	article := models.Article{
		Title:       title,
		Content:     content,
		URL:         "https://example.com/" + title,
		Author:      "Test Author",
		Source:      "Test Source",
		PublishedAt: time.Now(),
		Category:    "news",
		ImageURL:    "https://example.com/image.jpg",
	}

	err := repo.CreateArticle(&article)
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	// Get the created article to return its ID
	filters := models.ArticleFilters{Page: 1, Limit: 1}
	articles, _, err := repo.GetArticles(filters)
	if err != nil || len(articles) == 0 {
		t.Fatalf("Failed to get created article")
	}

	return articles[0].ID
}

func TestHealthHandler(t *testing.T) {
	handler := &Handler{}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.Health(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("handler returned wrong status: got %v want %v", response.Status, "ok")
	}

	if response.Version != "1.0.0" {
		t.Errorf("handler returned wrong version: got %v want %v", response.Version, "1.0.0")
	}
}

func TestGetArticles(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	// Create test articles
	createTestArticleForHandler(t, handler.articleRepo, "Article 1", "Content 1")
	createTestArticleForHandler(t, handler.articleRepo, "Article 2", "Content 2")

	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
		checkCount   bool
		expectedMin  int
	}{
		{
			name:         "Get all articles",
			queryParams:  "",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get articles with pagination",
			queryParams:  "?page=1&limit=1",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  1,
		},
		{
			name:         "Get articles by category",
			queryParams:  "?category=news",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Get articles by source",
			queryParams:  "?source=Test%20Source",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  2,
		},
		{
			name:         "Search articles",
			queryParams:  "?search=Article%201",
			expectedCode: http.StatusOK,
			checkCount:   true,
			expectedMin:  1,
		},
		{
			name:         "Invalid page parameter",
			queryParams:  "?page=invalid",
			expectedCode: http.StatusOK, // Should default to page 1
			checkCount:   false,
		},
		{
			name:         "Invalid limit parameter",
			queryParams:  "?limit=invalid",
			expectedCode: http.StatusOK, // Should default to 20
			checkCount:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/articles"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.GetArticles(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			if tt.checkCount && rr.Code == http.StatusOK {
				var response models.APIResponse[[]models.Article]
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				} else if !response.Success {
					t.Errorf("Expected success=true, got %v", response.Success)
				} else if len(response.Data) < tt.expectedMin {
					t.Errorf("Expected at least %d articles, got %d", tt.expectedMin, len(response.Data))
				}
			}
		})
	}
}

func TestGetArticle(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	// Create test article
	articleID := createTestArticleForHandler(t, handler.articleRepo, "Test Article", "Test Content")

	tests := []struct {
		name         string
		articleID    string
		expectedCode int
	}{
		{
			name:         "Valid article ID",
			articleID:    strconv.Itoa(articleID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid article ID",
			articleID:    "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Non-existent article ID",
			articleID:    "999",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/articles/"+tt.articleID, nil)
			rr := httptest.NewRecorder()

			// Add URL parameters to request context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.articleID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetArticle(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response models.APIResponse[models.Article]
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				} else if !response.Success {
					t.Errorf("Expected success=true, got %v", response.Success)
				} else if response.Data.Title != "Test Article" {
					t.Errorf("Expected article title 'Test Article', got %s", response.Data.Title)
				}
			}
		})
	}
}

func TestSummarizeArticle(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	// Create test article
	articleID := createTestArticleForHandler(t, handler.articleRepo, "Article to Summarize", "This is the first sentence. This is the second sentence. This is the third sentence.")

	tests := []struct {
		name         string
		articleID    string
		expectedCode int
	}{
		{
			name:         "Valid article ID",
			articleID:    strconv.Itoa(articleID),
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid article ID",
			articleID:    "invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Non-existent article ID",
			articleID:    "999",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/articles/"+tt.articleID+"/summarize", nil)
			rr := httptest.NewRecorder()

			// Add URL parameters to request context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.articleID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.SummarizeArticle(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response models.APIResponse[map[string]string]
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				} else if !response.Success {
					t.Errorf("Expected success=true, got %v", response.Success)
				} else if response.Data["summary"] == "" {
					t.Error("Expected non-empty summary")
				}
			}
		})
	}
}

func TestGetSources(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	// Create test source
	source := models.NewsSource{
		Name:     "Test Source",
		URL:      "https://test.com",
		FeedURL:  "https://test.com/rss",
		Category: "news",
		Active:   true,
	}
	err := handler.sourceRepo.CreateSource(&source)
	if err != nil {
		t.Fatalf("Failed to create test source: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	rr := httptest.NewRecorder()

	handler.GetSources(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.APIResponse[[]models.NewsSource]
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if len(response.Data) == 0 {
		t.Error("Expected at least one source")
	}
}

func TestGetCategories(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	// Create test articles with different categories
	createTestArticleForHandler(t, handler.articleRepo, "News Article", "News content")

	// Create article with different category
	article := models.Article{
		Title:       "Sports Article",
		Content:     "Sports content",
		URL:         "https://example.com/sports",
		Source:      "Sports Source",
		Category:    "sports",
		PublishedAt: time.Now(),
	}
	if err := handler.articleRepo.CreateArticle(&article); err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/categories", nil)
	rr := httptest.NewRecorder()

	handler.GetCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.APIResponse[[]string]
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if len(response.Data) == 0 {
		t.Error("Expected at least one category")
	}
}

func TestRefreshNews(t *testing.T) {
	db, handler := setupHandlerTest(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/v1/refresh", nil)
	rr := httptest.NewRecorder()

	handler.RefreshNews(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.APIResponse[map[string]string]
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	expectedMessage := "News refresh started"
	if response.Data["message"] != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, response.Data["message"])
	}
}

func TestParseArticleFilters(t *testing.T) {
	_, handler := setupHandlerTest(t)

	tests := []struct {
		name        string
		queryParams string
		expected    models.ArticleFilters
	}{
		{
			name:        "Default values",
			queryParams: "",
			expected:    models.ArticleFilters{Page: 1, Limit: 20},
		},
		{
			name:        "Custom pagination",
			queryParams: "?page=2&limit=10",
			expected:    models.ArticleFilters{Page: 2, Limit: 10},
		},
		{
			name:        "With filters",
			queryParams: "?category=news&source=Test&search=article",
			expected:    models.ArticleFilters{Page: 1, Limit: 20, Category: "news", Source: "Test", Search: "article"},
		},
		{
			name:        "Invalid page - should default",
			queryParams: "?page=0",
			expected:    models.ArticleFilters{Page: 1, Limit: 20},
		},
		{
			name:        "Invalid limit - should default",
			queryParams: "?limit=150", // Over 100 limit
			expected:    models.ArticleFilters{Page: 1, Limit: 20},
		},
		{
			name:        "Negative page - should default",
			queryParams: "?page=-1",
			expected:    models.ArticleFilters{Page: 1, Limit: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			filters, err := handler.parseArticleFilters(req)
			if err != nil {
				t.Errorf("parseArticleFilters returned error: %v", err)
			}

			if filters.Page != tt.expected.Page {
				t.Errorf("Expected page %d, got %d", tt.expected.Page, filters.Page)
			}
			if filters.Limit != tt.expected.Limit {
				t.Errorf("Expected limit %d, got %d", tt.expected.Limit, filters.Limit)
			}
			if filters.Category != tt.expected.Category {
				t.Errorf("Expected category %s, got %s", tt.expected.Category, filters.Category)
			}
			if filters.Source != tt.expected.Source {
				t.Errorf("Expected source %s, got %s", tt.expected.Source, filters.Source)
			}
			if filters.Search != tt.expected.Search {
				t.Errorf("Expected search %s, got %s", tt.expected.Search, filters.Search)
			}
		})
	}
}

func TestRespondJSON(t *testing.T) {
	_, handler := setupHandlerTest(t)

	rr := httptest.NewRecorder()
	testData := map[string]string{"test": "value"}

	handler.respondJSON(rr, http.StatusOK, testData)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["test"] != "value" {
		t.Errorf("Expected test=value, got test=%s", response["test"])
	}
}

func TestRespondError(t *testing.T) {
	_, handler := setupHandlerTest(t)

	rr := httptest.NewRecorder()
	handler.respondError(rr, http.StatusBadRequest, models.ErrCodeValidation, "Test error", "Test details")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse[any]
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Success != false {
		t.Errorf("Expected success=false, got %v", response.Success)
	}

	if response.Error == nil {
		t.Error("Expected error object, got nil")
		return
	}

	if response.Error.Message != "Test error" {
		t.Errorf("Expected error message 'Test error', got '%s'", response.Error.Message)
	}
	if response.Error.Code != models.ErrCodeValidation {
		t.Errorf("Expected error code '%s', got '%s'", models.ErrCodeValidation, response.Error.Code)
	}
	if response.Error.Details != "Test details" {
		t.Errorf("Expected error details 'Test details', got '%s'", response.Error.Details)
	}
}
