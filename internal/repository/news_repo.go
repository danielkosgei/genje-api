package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type NewsRepo struct {
	pool *pgxpool.Pool
}

func NewNewsRepo(pool *pgxpool.Pool) *NewsRepo {
	return &NewsRepo{pool: pool}
}

func (r *NewsRepo) ListArticles(ctx context.Context, f models.NewsFilter) ([]models.NewsArticle, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM news_articles na WHERE 1=1`
	dataQuery := `
		SELECT na.id, na.source_id, na.title, na.content, na.summary, na.url,
		       na.author, na.image_url, na.published_at, na.scraped_at,
		       na.category, na.is_election_related, na.created_at
		FROM news_articles na WHERE 1=1`

	var args []interface{}
	argIdx := 1
	where := ""

	if f.ElectionRelated != nil && *f.ElectionRelated {
		where += " AND na.is_election_related = true"
	}
	if f.SourceID != nil {
		where += fmt.Sprintf(" AND na.source_id = $%d", argIdx)
		args = append(args, *f.SourceID)
		argIdx++
	}
	if f.PoliticianID != nil {
		where += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM article_politician_mentions apm WHERE apm.article_id = na.id AND apm.politician_id = $%d)`, argIdx)
		args = append(args, *f.PoliticianID)
		argIdx++
	}
	if f.Since != nil {
		where += fmt.Sprintf(" AND na.published_at >= $%d", argIdx)
		args = append(args, *f.Since)
		argIdx++
	}
	if f.Until != nil {
		where += fmt.Sprintf(" AND na.published_at <= $%d", argIdx)
		args = append(args, *f.Until)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count articles: %w", err)
	}

	dataQuery += where + fmt.Sprintf(` ORDER BY na.published_at DESC NULLS LAST LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []models.NewsArticle
	for rows.Next() {
		var a models.NewsArticle
		if err := rows.Scan(
			&a.ID, &a.SourceID, &a.Title, &a.Content, &a.Summary, &a.URL,
			&a.Author, &a.ImageURL, &a.PublishedAt, &a.ScrapedAt,
			&a.Category, &a.IsElectionRelated, &a.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, total, nil
}

func (r *NewsRepo) GetArticleByID(ctx context.Context, id uuid.UUID) (*models.NewsArticle, error) {
	query := `
		SELECT id, source_id, title, content, summary, url, author, image_url,
		       published_at, scraped_at, category, is_election_related, created_at
		FROM news_articles WHERE id = $1`

	var a models.NewsArticle
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.SourceID, &a.Title, &a.Content, &a.Summary, &a.URL,
		&a.Author, &a.ImageURL, &a.PublishedAt, &a.ScrapedAt,
		&a.Category, &a.IsElectionRelated, &a.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}
	return &a, nil
}

func (r *NewsRepo) GetArticlesByPolitician(ctx context.Context, politicianID uuid.UUID, limit, offset int) ([]models.NewsArticle, int, error) {
	f := models.NewsFilter{
		PoliticianID: &politicianID,
		Limit:        limit,
		Offset:       offset,
	}
	return r.ListArticles(ctx, f)
}

func (r *NewsRepo) ListSources(ctx context.Context) ([]models.NewsSource, error) {
	query := `
		SELECT id, name, url, feed_url, type, outlet, active, created_at, updated_at
		FROM news_sources ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.FeedURL, &s.Type, &s.Outlet, &s.Active, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (r *NewsRepo) GetActiveSources(ctx context.Context) ([]models.NewsSource, error) {
	query := `
		SELECT id, name, url, feed_url, type, outlet, active, created_at, updated_at
		FROM news_sources WHERE active = true ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get active sources: %w", err)
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.FeedURL, &s.Type, &s.Outlet, &s.Active, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (r *NewsRepo) InsertArticle(ctx context.Context, a *models.NewsArticle) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO news_articles (source_id, title, content, summary, url, author, image_url, published_at, scraped_at, category, is_election_related)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9, $10)
		 ON CONFLICT (url) DO NOTHING`,
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
