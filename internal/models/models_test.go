package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestArticleJSONMarshaling(t *testing.T) {
	publishedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC)

	article := Article{
		ID:          1,
		Title:       "Test Article",
		Content:     "This is test content",
		Summary:     "Test summary",
		URL:         "https://example.com/article/1",
		Author:      "Test Author",
		Source:      "Test Source",
		PublishedAt: publishedAt,
		CreatedAt:   createdAt,
		Category:    "technology",
		ImageURL:    "https://example.com/image.jpg",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("Failed to marshal article to JSON: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled Article
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal article from JSON: %v", err)
	}

	// Compare fields
	if unmarshaled.ID != article.ID {
		t.Errorf("ID mismatch: expected %d, got %d", article.ID, unmarshaled.ID)
	}
	if unmarshaled.Title != article.Title {
		t.Errorf("Title mismatch: expected %s, got %s", article.Title, unmarshaled.Title)
	}
	if unmarshaled.URL != article.URL {
		t.Errorf("URL mismatch: expected %s, got %s", article.URL, unmarshaled.URL)
	}
}

func TestNewsSourceJSONMarshaling(t *testing.T) {
	source := NewsSource{
		ID:       1,
		Name:     "Test News Source",
		URL:      "https://testnews.com",
		FeedURL:  "https://testnews.com/rss",
		Category: "news",
		Active:   true,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("Failed to marshal source to JSON: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled NewsSource
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal source from JSON: %v", err)
	}

	// Compare fields
	if unmarshaled.ID != source.ID {
		t.Errorf("ID mismatch: expected %d, got %d", source.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != source.Name {
		t.Errorf("Name mismatch: expected %s, got %s", source.Name, unmarshaled.Name)
	}
	if unmarshaled.Active != source.Active {
		t.Errorf("Active mismatch: expected %v, got %v", source.Active, unmarshaled.Active)
	}
}

func TestArticleFiltersDefaults(t *testing.T) {
	filters := ArticleFilters{}

	// Test default values
	if filters.Page != 0 {
		t.Errorf("Expected default Page to be 0, got %d", filters.Page)
	}
	if filters.Limit != 0 {
		t.Errorf("Expected default Limit to be 0, got %d", filters.Limit)
	}
	if filters.Category != "" {
		t.Errorf("Expected default Category to be empty, got %s", filters.Category)
	}
	if filters.Source != "" {
		t.Errorf("Expected default Source to be empty, got %s", filters.Source)
	}
	if filters.Search != "" {
		t.Errorf("Expected default Search to be empty, got %s", filters.Search)
	}
}

func TestPaginationResponse(t *testing.T) {
	pagination := PaginationResponse{
		Page:  1,
		Limit: 20,
		Total: 100,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(pagination)
	if err != nil {
		t.Fatalf("Failed to marshal pagination to JSON: %v", err)
	}

	var unmarshaled PaginationResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal pagination from JSON: %v", err)
	}

	if unmarshaled.Page != pagination.Page {
		t.Errorf("Page mismatch: expected %d, got %d", pagination.Page, unmarshaled.Page)
	}
	if unmarshaled.Limit != pagination.Limit {
		t.Errorf("Limit mismatch: expected %d, got %d", pagination.Limit, unmarshaled.Limit)
	}
	if unmarshaled.Total != pagination.Total {
		t.Errorf("Total mismatch: expected %d, got %d", pagination.Total, unmarshaled.Total)
	}
}

func TestArticlesResponse(t *testing.T) {
	articles := []Article{
		{ID: 1, Title: "Article 1", URL: "https://example.com/1", Source: "Source 1"},
		{ID: 2, Title: "Article 2", URL: "https://example.com/2", Source: "Source 2"},
	}

	pagination := PaginationResponse{
		Page:  1,
		Limit: 2,
		Total: 2,
	}

	response := ArticlesResponse{
		Articles:   articles,
		Pagination: pagination,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal articles response to JSON: %v", err)
	}

	var unmarshaled ArticlesResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal articles response from JSON: %v", err)
	}

	if len(unmarshaled.Articles) != len(response.Articles) {
		t.Errorf("Articles count mismatch: expected %d, got %d", len(response.Articles), len(unmarshaled.Articles))
	}

	if unmarshaled.Pagination.Total != response.Pagination.Total {
		t.Errorf("Pagination total mismatch: expected %d, got %d", response.Pagination.Total, unmarshaled.Pagination.Total)
	}
}

func TestHealthResponse(t *testing.T) {
	now := time.Now()
	health := HealthResponse{
		Status:    "ok",
		Timestamp: now,
		Version:   "1.0.0",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("Failed to marshal health response to JSON: %v", err)
	}

	var unmarshaled HealthResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal health response from JSON: %v", err)
	}

	if unmarshaled.Status != health.Status {
		t.Errorf("Status mismatch: expected %s, got %s", health.Status, unmarshaled.Status)
	}
	if unmarshaled.Version != health.Version {
		t.Errorf("Version mismatch: expected %s, got %s", health.Version, unmarshaled.Version)
	}
}

func TestErrorResponse(t *testing.T) {
	errorResp := ErrorResponse{
		Error:   "Test error",
		Code:    400,
		Details: "Test error details",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response to JSON: %v", err)
	}

	var unmarshaled ErrorResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response from JSON: %v", err)
	}

	if unmarshaled.Error != errorResp.Error {
		t.Errorf("Error mismatch: expected %s, got %s", errorResp.Error, unmarshaled.Error)
	}
	if unmarshaled.Code != errorResp.Code {
		t.Errorf("Code mismatch: expected %d, got %d", errorResp.Code, unmarshaled.Code)
	}
	if unmarshaled.Details != errorResp.Details {
		t.Errorf("Details mismatch: expected %s, got %s", errorResp.Details, unmarshaled.Details)
	}
}

func TestSourcesResponse(t *testing.T) {
	sources := []NewsSource{
		{ID: 1, Name: "Source 1", URL: "https://source1.com", Active: true},
		{ID: 2, Name: "Source 2", URL: "https://source2.com", Active: false},
	}

	response := SourcesResponse{Sources: sources}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal sources response to JSON: %v", err)
	}

	var unmarshaled SourcesResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal sources response from JSON: %v", err)
	}

	if len(unmarshaled.Sources) != len(response.Sources) {
		t.Errorf("Sources count mismatch: expected %d, got %d", len(response.Sources), len(unmarshaled.Sources))
	}
}

func TestCategoriesResponse(t *testing.T) {
	categories := []string{"news", "sports", "technology", "business"}
	response := CategoriesResponse{Categories: categories}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal categories response to JSON: %v", err)
	}

	var unmarshaled CategoriesResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal categories response from JSON: %v", err)
	}

	if len(unmarshaled.Categories) != len(response.Categories) {
		t.Errorf("Categories count mismatch: expected %d, got %d", len(response.Categories), len(unmarshaled.Categories))
	}

	for i, category := range response.Categories {
		if unmarshaled.Categories[i] != category {
			t.Errorf("Category %d mismatch: expected %s, got %s", i, category, unmarshaled.Categories[i])
		}
	}
}

func TestArticleValidation(t *testing.T) {
	// Test article with minimal required fields
	article := Article{
		Title:  "Test Title",
		URL:    "https://example.com",
		Source: "Test Source",
	}

	if article.Title == "" {
		t.Error("Article should have a title")
	}
	if article.URL == "" {
		t.Error("Article should have a URL")
	}
	if article.Source == "" {
		t.Error("Article should have a source")
	}
}

func TestNewsSourceValidation(t *testing.T) {
	// Test source with minimal required fields
	source := NewsSource{
		Name:    "Test Source",
		URL:     "https://example.com",
		FeedURL: "https://example.com/rss",
	}

	if source.Name == "" {
		t.Error("NewsSource should have a name")
	}
	if source.URL == "" {
		t.Error("NewsSource should have a URL")
	}
	if source.FeedURL == "" {
		t.Error("NewsSource should have a FeedURL")
	}
} 