package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"genje-api/internal/models"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) GetArticles(filters models.ArticleFilters) ([]models.Article, int, error) {
	// Build query with filters - use COALESCE to handle NULL values
	query := `SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
		COALESCE(author, ''), source, published_at, created_at, 
		COALESCE(category, 'general'), COALESCE(image_url, '') 
		FROM articles WHERE 1=1`
	countQuery := "SELECT COUNT(*) FROM articles WHERE 1=1"
	args := []interface{}{}

	if filters.Category != "" {
		query += " AND category = ?"
		countQuery += " AND category = ?"
		args = append(args, filters.Category)
	}

	if filters.Source != "" {
		query += " AND source = ?"
		countQuery += " AND source = ?"
		args = append(args, filters.Source)
	}

	if filters.Search != "" {
		query += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		countQuery += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Add pagination
	query += " ORDER BY published_at DESC LIMIT ? OFFSET ?"
	offset := (filters.Page - 1) * filters.Limit
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return articles, total, nil
}

func (r *ArticleRepository) GetArticleByID(id int) (*models.Article, error) {
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles WHERE id = ?
	`

	var article models.Article
	err := r.db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Content,
		&article.Summary, &article.URL, &article.Author, &article.Source, &article.PublishedAt,
		&article.CreatedAt, &article.Category, &article.ImageURL)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &article, nil
}

func (r *ArticleRepository) CreateArticle(article *models.Article) error {
	query := `
		INSERT OR IGNORE INTO articles (title, content, url, author, source, published_at, category, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Ensure we never insert NULL values for string fields
	content := article.Content
	if content == "" {
		content = ""
	}
	author := article.Author
	if author == "" {
		author = ""
	}
	category := article.Category
	if category == "" {
		category = "general"
	}
	imageURL := article.ImageURL
	if imageURL == "" {
		imageURL = ""
	}

	_, err := r.db.Exec(query, article.Title, content, article.URL, author,
		article.Source, article.PublishedAt, category, imageURL)

	if err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) UpdateSummary(id int, summary string) error {
	query := "UPDATE articles SET summary = ? WHERE id = ?"
	_, err := r.db.Exec(query, summary, id)
	if err != nil {
		return fmt.Errorf("failed to update summary: %w", err)
	}
	return nil
}

func (r *ArticleRepository) GetCategories() ([]string, error) {
	query := "SELECT DISTINCT category FROM articles ORDER BY category"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			continue
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *ArticleRepository) CreateArticlesBatch(articles []models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error as we're already in error handling
	}()

	query := `
		INSERT OR IGNORE INTO articles (title, content, url, author, source, published_at, category, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, article := range articles {
		// Ensure we never insert NULL values for string fields
		content := article.Content
		if content == "" {
			content = ""
		}
		author := article.Author
		if author == "" {
			author = ""
		}
		category := article.Category
		if category == "" {
			category = "general"
		}
		imageURL := article.ImageURL
		if imageURL == "" {
			imageURL = ""
		}

		_, err := stmt.Exec(article.Title, content, article.URL, author,
			article.Source, article.PublishedAt, category, imageURL)
		if err != nil {
			return fmt.Errorf("failed to insert article: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetGlobalStats returns global statistics about articles
func (r *ArticleRepository) GetGlobalStats() (models.GlobalStats, error) {
	var stats models.GlobalStats
	
	query := `
		SELECT 
			COUNT(*) as total_articles,
			COUNT(DISTINCT category) as categories
		FROM articles
	`
	
	err := r.db.QueryRow(query).Scan(&stats.TotalArticles, &stats.Categories)
	if err != nil {
		return stats, fmt.Errorf("failed to get global stats: %w", err)
	}
	
	// Get last updated time separately
	var lastUpdatedStr string
	err = r.db.QueryRow("SELECT MAX(created_at) FROM articles").Scan(&lastUpdatedStr)
	if err != nil {
		// If no articles, use current time
		stats.LastUpdated = time.Now()
	} else {
		// Parse the SQLite datetime string
		stats.LastUpdated, err = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		if err != nil {
			// If parsing fails, use current time
			stats.LastUpdated = time.Now()
		}
	}
	
	return stats, nil
}

// GetSourceStats returns statistics per source
func (r *ArticleRepository) GetSourceStats() ([]models.SourceStats, error) {
	query := `
		SELECT 
			source,
			COUNT(*) as article_count,
			COALESCE(category, 'general') as category,
			MAX(created_at) as last_updated
		FROM articles 
		GROUP BY source, category
		ORDER BY article_count DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}
	defer rows.Close()
	
	var stats []models.SourceStats
	for rows.Next() {
		var stat models.SourceStats
		var lastUpdatedStr string
		err := rows.Scan(&stat.Name, &stat.ArticleCount, &stat.Category, &lastUpdatedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source stats: %w", err)
		}
		
		// Parse the SQLite datetime string
		stat.LastUpdated, err = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		if err != nil {
			// If parsing fails, use current time
			stat.LastUpdated = time.Now()
		}
		
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// GetCategoryStats returns statistics per category
func (r *ArticleRepository) GetCategoryStats() ([]models.CategoryStats, error) {
	query := `
		SELECT 
			COALESCE(category, 'general') as category,
			COUNT(*) as article_count
		FROM articles 
		GROUP BY category
		ORDER BY article_count DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()
	
	var stats []models.CategoryStats
	for rows.Next() {
		var stat models.CategoryStats
		err := rows.Scan(&stat.Category, &stat.ArticleCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category stats: %w", err)
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// GetTimelineStats returns article count over time (last 30 days)
func (r *ArticleRepository) GetTimelineStats(days int) ([]models.TimelineStats, error) {
	if days <= 0 {
		days = 30
	}
	
	query := `
		SELECT 
			DATE(published_at) as date,
			COUNT(*) as article_count
		FROM articles 
		WHERE published_at >= DATE('now', '-' || ? || ' days')
		GROUP BY DATE(published_at)
		ORDER BY date DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, days, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline stats: %w", err)
	}
	defer rows.Close()
	
	var stats []models.TimelineStats
	for rows.Next() {
		var stat models.TimelineStats
		err := rows.Scan(&stat.Date, &stat.ArticleCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline stats: %w", err)
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// GetRecentArticles returns articles from the last N hours
func (r *ArticleRepository) GetRecentArticles(hours int, limit int) ([]models.Article, error) {
	if hours <= 0 {
		hours = 24
	}
	if limit <= 0 {
		limit = 20
	}
	
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles 
		WHERE published_at >= DATETIME('now', '-' || ? || ' hours')
		ORDER BY published_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent articles: %w", err)
	}
	defer rows.Close()
	
	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent article: %w", err)
		}
		articles = append(articles, article)
	}
	
	return articles, nil
}

// GetSystemStatus returns system status information
func (r *ArticleRepository) GetSystemStatus() (models.SystemStatus, error) {
	var status models.SystemStatus
	status.Status = "ok"
	
	// Get total articles
	err := r.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&status.TotalArticles)
	if err != nil {
		return status, fmt.Errorf("failed to get total articles: %w", err)
	}
	
	// Get last aggregation time (most recent article creation)
	var lastAggregationStr string
	err = r.db.QueryRow("SELECT MAX(created_at) FROM articles").Scan(&lastAggregationStr)
	if err != nil {
		// If no articles, set to current time
		status.LastAggregation = time.Now()
	} else {
		// Parse the SQLite datetime string
		status.LastAggregation, err = time.Parse("2006-01-02 15:04:05", lastAggregationStr)
		if err != nil {
			// If parsing fails, use current time
			status.LastAggregation = time.Now()
		}
	}
	
	return status, nil
}

// SearchArticles performs full-text search on articles
func (r *ArticleRepository) SearchArticles(filters models.SearchFilters) ([]models.Article, int, error) {
	// Build search query with ranking
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM articles WHERE 1=1"
	args := []interface{}{}

	// Search in title and content
	if filters.Query != "" {
		query += " AND (title LIKE ? OR COALESCE(content, '') LIKE ? OR COALESCE(summary, '') LIKE ?)"
		countQuery += " AND (title LIKE ? OR COALESCE(content, '') LIKE ? OR COALESCE(summary, '') LIKE ?)"
		searchTerm := "%" + filters.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	// Filter by category
	if filters.Category != "" {
		query += " AND category = ?"
		countQuery += " AND category = ?"
		args = append(args, filters.Category)
	}

	// Filter by source
	if filters.Source != "" {
		query += " AND source = ?"
		countQuery += " AND source = ?"
		args = append(args, filters.Source)
	}

	// Date range filters
	if filters.From != "" {
		query += " AND published_at >= ?"
		countQuery += " AND published_at >= ?"
		args = append(args, filters.From)
	}

	if filters.To != "" {
		query += " AND published_at <= ?"
		countQuery += " AND published_at <= ?"
		args = append(args, filters.To)
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get search count: %w", err)
	}

	// Add sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "relevance"
	}
	
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	
	switch sortBy {
	case "date":
		query += " ORDER BY published_at " + strings.ToUpper(sortOrder)
	case "source":
		query += " ORDER BY source " + strings.ToUpper(sortOrder)
	default: // relevance
		query += " ORDER BY published_at " + strings.ToUpper(sortOrder)
	}

	// Add pagination
	query += " LIMIT ? OFFSET ?"
	offset := (filters.Page - 1) * filters.Limit
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search articles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, total, nil
}

// GetArticlesFeed returns paginated articles with cursor-based pagination
func (r *ArticleRepository) GetArticlesFeed(req models.FeedRequest) (models.CursorPaginationResponse, error) {
	var result models.CursorPaginationResponse
	
	// Build base query
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if req.Category != "" {
		query += " AND category = ?"
		args = append(args, req.Category)
	}

	if req.Source != "" {
		query += " AND source = ?"
		args = append(args, req.Source)
	}

	// Handle cursor pagination
	if req.Cursor != "" {
		// Decode cursor (simplified - in production, use proper cursor encoding)
		query += " AND id > ?"
		args = append(args, req.Cursor)
	}

	// Sorting
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "date"
	}
	
	switch sortBy {
	case "popularity":
		query += " ORDER BY published_at DESC, id DESC"
	default: // date
		query += " ORDER BY published_at DESC, id DESC"
	}

	// Add limit + 1 to check if there are more results
	query += " LIMIT ?"
	args = append(args, req.Limit+1)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return result, fmt.Errorf("failed to get articles feed: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return result, fmt.Errorf("failed to scan feed article: %w", err)
		}
		articles = append(articles, article)
	}

	// Check if there are more results
	result.HasMore = len(articles) > req.Limit
	if result.HasMore {
		articles = articles[:req.Limit] // Remove the extra article
		// Set next cursor to the last article's ID
		result.NextCursor = fmt.Sprintf("%d", articles[len(articles)-1].ID)
	}

	result.Articles = articles
	result.Total = len(articles)

	return result, nil
}

// GetArticlesBySource returns articles from a specific source
func (r *ArticleRepository) GetArticlesBySource(sourceID int, filters models.ArticleFilters) ([]models.Article, int, error) {
	// First get the source name from the sources table
	var sourceName string
	err := r.db.QueryRow("SELECT name FROM sources WHERE id = ?", sourceID).Scan(&sourceName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get source name: %w", err)
	}

	// Now get articles from this source
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles WHERE source = ?
	`
	countQuery := "SELECT COUNT(*) FROM articles WHERE source = ?"
	args := []interface{}{sourceName}

	// Apply additional filters
	if filters.Category != "" {
		query += " AND category = ?"
		countQuery += " AND category = ?"
		args = append(args, filters.Category)
	}

	if filters.Search != "" {
		query += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		countQuery += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get source articles count: %w", err)
	}

	// Add pagination
	query += " ORDER BY published_at DESC LIMIT ? OFFSET ?"
	offset := (filters.Page - 1) * filters.Limit
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get articles by source: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan source article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, total, nil
}

// GetArticlesByCategory returns articles from a specific category
func (r *ArticleRepository) GetArticlesByCategory(category string, filters models.ArticleFilters) ([]models.Article, int, error) {
	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles WHERE category = ?
	`
	countQuery := "SELECT COUNT(*) FROM articles WHERE category = ?"
	args := []interface{}{category}

	// Apply additional filters
	if filters.Source != "" {
		query += " AND source = ?"
		countQuery += " AND source = ?"
		args = append(args, filters.Source)
	}

	if filters.Search != "" {
		query += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		countQuery += " AND (title LIKE ? OR COALESCE(content, '') LIKE ?)"
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get category articles count: %w", err)
	}

	// Add pagination
	query += " ORDER BY published_at DESC LIMIT ? OFFSET ?"
	offset := (filters.Page - 1) * filters.Limit
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get articles by category: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan category article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, total, nil
}

// GetTrendingArticles returns trending articles (simplified algorithm)
func (r *ArticleRepository) GetTrendingArticles(limit int, timeWindow string) ([]models.TrendingArticle, error) {
	if limit <= 0 {
		limit = 20
	}

	// Simplified trending algorithm based on recent articles
	hours := 24
	switch timeWindow {
	case "1h":
		hours = 1
	case "6h":
		hours = 6
	case "12h":
		hours = 12
	case "24h":
		hours = 24
	case "7d":
		hours = 24 * 7
	}

	query := `
		SELECT id, title, COALESCE(content, ''), COALESCE(summary, ''), url, 
			COALESCE(author, ''), source, published_at, created_at, 
			COALESCE(category, 'general'), COALESCE(image_url, '')
		FROM articles 
		WHERE published_at >= DATETIME('now', '-' || ? || ' hours')
		ORDER BY published_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending articles: %w", err)
	}
	defer rows.Close()

	var trendingArticles []models.TrendingArticle
	for rows.Next() {
		var trending models.TrendingArticle
		err := rows.Scan(&trending.ID, &trending.Title, &trending.Content, &trending.Summary,
			&trending.URL, &trending.Author, &trending.Source, &trending.PublishedAt,
			&trending.CreatedAt, &trending.Category, &trending.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending article: %w", err)
		}
		
		// Calculate simple trending score (mock data for now)
		trending.Score = 0.8 + (0.2 * float64(trending.ID%100)/100)
		trending.ViewCount = trending.ID%1000 + 50
		trending.ShareCount = trending.ID%100 + 10
		trending.CommentCount = trending.ID%50 + 5
		trending.TrendingReason = "Recent publication"
		
		trendingArticles = append(trendingArticles, trending)
	}

	return trendingArticles, nil
}

// GetTrendingTopics returns trending topics/keywords (simplified)
func (r *ArticleRepository) GetTrendingTopics(limit int, timeWindow string) ([]models.TrendingTopic, error) {
	if limit <= 0 {
		limit = 10
	}

	// Simplified trending topics based on frequent words in titles
	hours := 24
	switch timeWindow {
	case "1h":
		hours = 1
	case "6h":
		hours = 6
	case "12h":
		hours = 12
	case "24h":
		hours = 24
	case "7d":
		hours = 24 * 7
	}

	// Get most frequent categories as trending topics
	query := `
		SELECT 
			category as topic,
			COUNT(*) as count,
			category,
			MIN(published_at) as first_seen,
			MAX(published_at) as last_updated
		FROM articles 
		WHERE published_at >= DATETIME('now', '-' || ? || ' hours')
		GROUP BY category
		ORDER BY count DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending topics: %w", err)
	}
	defer rows.Close()

	var topics []models.TrendingTopic
	for rows.Next() {
		var topic models.TrendingTopic
		var firstSeenStr, lastUpdatedStr string
		
		err := rows.Scan(&topic.Topic, &topic.Count, &topic.Category, 
			&firstSeenStr, &lastUpdatedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending topic: %w", err)
		}
		
		// Parse dates
		topic.FirstSeen, _ = time.Parse("2006-01-02 15:04:05", firstSeenStr)
		topic.LastUpdated, _ = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		
		// Calculate simple score
		topic.Score = float64(topic.Count) * 0.1
		
		// Get sample article IDs for this topic
		articleQuery := `
			SELECT id FROM articles 
			WHERE category = ? AND published_at >= DATETIME('now', '-' || ? || ' hours')
			ORDER BY published_at DESC
			LIMIT 5
		`
		articleRows, err := r.db.Query(articleQuery, topic.Category, hours)
		if err == nil {
			var articleIDs []int
			for articleRows.Next() {
				var id int
				if articleRows.Scan(&id) == nil {
					articleIDs = append(articleIDs, id)
				}
			}
			topic.Articles = articleIDs
			articleRows.Close()
		}
		
		topics = append(topics, topic)
	}

	return topics, nil
} 