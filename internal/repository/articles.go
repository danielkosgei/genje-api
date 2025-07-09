package repository

import (
	"database/sql"
	"fmt"

	"genje-api/internal/models"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) GetArticles(filters models.ArticleFilters) ([]models.Article, int, error) {
	// Build query with filters
	query := "SELECT id, title, content, summary, url, author, source, published_at, created_at, category, image_url FROM articles WHERE 1=1"
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
		query += " AND (title LIKE ? OR content LIKE ?)"
		countQuery += " AND (title LIKE ? OR content LIKE ?)"
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
		SELECT id, title, content, summary, url, author, source, published_at, created_at, category, image_url
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

	_, err := r.db.Exec(query, article.Title, article.Content, article.URL, article.Author,
		article.Source, article.PublishedAt, article.Category, article.ImageURL)

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
		_, err := stmt.Exec(article.Title, article.Content, article.URL, article.Author,
			article.Source, article.PublishedAt, article.Category, article.ImageURL)
		if err != nil {
			return fmt.Errorf("failed to insert article: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
} 