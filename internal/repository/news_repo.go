package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/category"
	"jalada/internal/models"
)

// articleColumns is the shared SELECT list for articles. The publisher name comes from the source
// join so clients get something readable without a second request.
const articleColumns = `
	na.id, na.source_id, COALESCE(ns.publisher, ns.name), na.title, na.content, na.summary, na.url,
	na.author, na.image_url, na.published_at, na.scraped_at,
	na.category, na.is_election_related, na.created_at`

const articleFrom = ` news_articles na LEFT JOIN news_sources ns ON ns.id = na.source_id`

type NewsRepo struct {
	pool *pgxpool.Pool
}

func NewNewsRepo(pool *pgxpool.Pool) *NewsRepo {
	return &NewsRepo{pool: pool}
}

// ListArticles returns one page of articles that match f, newest first, and the total number of matches.
func (r *NewsRepo) ListArticles(ctx context.Context, f models.NewsFilter) ([]models.NewsArticle, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	where, args := buildArticleFilter(f)

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM`+articleFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count articles: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	dataQuery := `SELECT ` + articleColumns + ` FROM` + articleFrom + where +
		fmt.Sprintf(` ORDER BY na.published_at DESC NULLS LAST, na.id LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []models.NewsArticle
	for rows.Next() {
		var a models.NewsArticle
		if err := scanArticle(rows, &a); err != nil {
			return nil, 0, err
		}
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate articles: %w", err)
	}
	return articles, total, nil
}

// buildArticleFilter returns the WHERE clause for f and its positional arguments.
func buildArticleFilter(f models.NewsFilter) (string, []interface{}) {
	var b strings.Builder
	var args []interface{}
	add := func(clause string, value interface{}) {
		args = append(args, value)
		b.WriteString(fmt.Sprintf(" AND "+clause, len(args)))
	}

	b.WriteString(" WHERE TRUE")
	if f.ElectionRelated != nil && *f.ElectionRelated {
		b.WriteString(" AND na.is_election_related = true")
	}
	if f.SourceID != nil {
		add("na.source_id = $%d", *f.SourceID)
	}
	if f.PoliticianID != nil {
		add("EXISTS (SELECT 1 FROM article_politician_mentions apm WHERE apm.article_id = na.id AND apm.politician_id = $%d)", *f.PoliticianID)
	}
	if f.Category != nil && *f.Category != "" {
		add("LOWER(na.category) = LOWER($%d)", *f.Category)
	}
	if f.Query != "" {
		pattern := "%" + escapeLike(f.Query) + "%"
		args = append(args, pattern)
		n := len(args)
		b.WriteString(fmt.Sprintf(
			" AND (na.title ILIKE $%d OR na.summary ILIKE $%d OR na.content ILIKE $%d)", n, n, n))
	}
	if f.Since != nil {
		add("na.published_at >= $%d", *f.Since)
	}
	if f.Until != nil {
		add("na.published_at <= $%d", *f.Until)
	}
	return b.String(), args
}

// escapeLike escapes the LIKE wildcards so a query is matched literally.
func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}

func (r *NewsRepo) GetArticleByID(ctx context.Context, id uuid.UUID) (*models.NewsArticle, error) {
	query := `SELECT ` + articleColumns + ` FROM` + articleFrom + ` WHERE na.id = $1`

	var a models.NewsArticle
	err := scanArticle(r.pool.QueryRow(ctx, query, id), &a)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanArticle(row rowScanner, a *models.NewsArticle) error {
	if err := row.Scan(
		&a.ID, &a.SourceID, &a.Source, &a.Title, &a.Content, &a.Summary, &a.URL,
		&a.Author, &a.ImageURL, &a.PublishedAt, &a.ScrapedAt,
		&a.Category, &a.IsElectionRelated, &a.CreatedAt,
	); err != nil {
		return fmt.Errorf("scan article: %w", err)
	}
	return nil
}

func (r *NewsRepo) GetArticlesByPolitician(ctx context.Context, politicianID uuid.UUID, limit, offset int) ([]models.NewsArticle, int, error) {
	f := models.NewsFilter{
		PoliticianID: &politicianID,
		Limit:        limit,
		Offset:       offset,
	}
	return r.ListArticles(ctx, f)
}

// BackfillCategories re-derives the category of every stored article using the current rules.
// It is safe to run on every boot: articles whose category is already right are not touched.
// It returns how many articles changed.
func (r *NewsRepo) BackfillCategories(ctx context.Context) (int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT na.id, na.url, na.title, na.category, ns.default_category
		FROM news_articles na LEFT JOIN news_sources ns ON ns.id = na.source_id`)
	if err != nil {
		return 0, fmt.Errorf("list articles for backfill: %w", err)
	}

	type update struct {
		id       uuid.UUID
		category string
	}
	var pending []update
	for rows.Next() {
		var (
			id              uuid.UUID
			url, title      string
			label, fallback *string
		)
		if err := rows.Scan(&id, &url, &title, &label, &fallback); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan article for backfill: %w", err)
		}
		canonical := category.Normalize(deref(label), url, title, deref(fallback))
		if label == nil || *label != canonical {
			pending = append(pending, update{id: id, category: canonical})
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate articles for backfill: %w", err)
	}

	for _, u := range pending {
		if _, err := r.pool.Exec(ctx, `UPDATE news_articles SET category = $2 WHERE id = $1`, u.id, u.category); err != nil {
			return 0, fmt.Errorf("update article category: %w", err)
		}
	}
	return len(pending), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *NewsRepo) ListSources(ctx context.Context) ([]models.NewsSource, error) {
	return r.querySources(ctx, `
		SELECT id, name, url, feed_url, type, outlet, publisher, default_category, active, created_at, updated_at
		FROM news_sources ORDER BY name`)
}

func (r *NewsRepo) GetActiveSources(ctx context.Context) ([]models.NewsSource, error) {
	return r.querySources(ctx, `
		SELECT id, name, url, feed_url, type, outlet, publisher, default_category, active, created_at, updated_at
		FROM news_sources WHERE active = true ORDER BY name`)
}

func (r *NewsRepo) querySources(ctx context.Context, query string) ([]models.NewsSource, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list news sources: %w", err)
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.FeedURL, &s.Type, &s.Outlet, &s.Publisher, &s.DefaultCategory, &s.Active, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan news source: %w", err)
		}
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate news sources: %w", err)
	}
	return sources, nil
}

func (r *NewsRepo) InsertArticle(ctx context.Context, a *models.NewsArticle) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO news_articles (source_id, title, content, summary, url, author, image_url, published_at, scraped_at, category, is_election_related)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9, $10)
		 ON CONFLICT (url) DO UPDATE SET image_url = COALESCE(EXCLUDED.image_url, news_articles.image_url)`,
		a.SourceID, a.Title, a.Content, a.Summary, a.URL, a.Author, a.ImageURL, a.PublishedAt, a.Category, a.IsElectionRelated,
	)
	if err != nil {
		return fmt.Errorf("insert article: %w", err)
	}
	return nil
}

func (r *NewsRepo) ArticleExistsByURL(ctx context.Context, url string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM news_articles WHERE url = $1)`, url).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check article exists: %w", err)
	}
	return exists, nil
}

func (r *NewsRepo) InsertMention(ctx context.Context, articleURL string, politicianID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO article_politician_mentions (article_id, politician_id)
		 SELECT na.id, $2
		 FROM news_articles na WHERE na.url = $1
		 ON CONFLICT (article_id, politician_id) DO NOTHING`,
		articleURL, politicianID,
	)
	if err != nil {
		return fmt.Errorf("insert mention: %w", err)
	}
	return nil
}

func (r *NewsRepo) GetAllPoliticianNames(ctx context.Context) ([]models.PoliticianSummary, error) {
	query := `SELECT id, slug, first_name, last_name, photo_url FROM politicians`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get politician names: %w", err)
	}
	defer rows.Close()

	var politicians []models.PoliticianSummary
	for rows.Next() {
		var p models.PoliticianSummary
		if err := rows.Scan(&p.ID, &p.Slug, &p.FirstName, &p.LastName, &p.PhotoURL); err != nil {
			return nil, fmt.Errorf("scan politician: %w", err)
		}
		politicians = append(politicians, p)
	}
	return politicians, nil
}

func (r *NewsRepo) ListDataSources(ctx context.Context) ([]models.Source, error) {
	query := `
		SELECT id, name, url, type, reliability, last_accessed_at, created_at, updated_at
		FROM sources ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list data sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var s models.Source
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Type, &s.Reliability, &s.LastAccessedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}
