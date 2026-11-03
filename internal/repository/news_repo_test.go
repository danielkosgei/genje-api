package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/database"
	"jalada/internal/models"
)

// testPool connects to DATABASE_URL and applies migrations. The test is skipped when the variable is unset.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping database integration test")
	}
	if err := database.RunMigrations(url); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedTestArticles inserts articles under a test:// URL prefix and removes them when the test ends.
func seedTestArticles(t *testing.T, pool *pgxpool.Pool, articles ...models.NewsArticle) {
	t.Helper()
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM news_articles WHERE url LIKE 'test://%'`)
	})
	for i := range articles {
		a := &articles[i]
		if err := NewNewsRepo(pool).InsertArticle(ctx, a); err != nil {
			t.Fatalf("insert %s: %v", a.Title, err)
		}
	}
}

func str(s string) *string { return &s }

func TestListArticlesFilters(t *testing.T) {
	pool := testPool(t)
	repo := NewNewsRepo(pool)
	seedTestArticles(t, pool,
		models.NewsArticle{URL: "test://1", Title: "Ruto speaks on budget", Category: str("Politics")},
		models.NewsArticle{URL: "test://2", Title: "Cup final tonight", Category: str("Sports")},
		models.NewsArticle{URL: "test://3", Title: "Budget day in pictures", Summary: str("A look at the ruto budget"), Category: str("Business")},
		models.NewsArticle{URL: "test://4", Title: "100% sure of it", Category: str("News")},
	)
	ctx := context.Background()

	tests := []struct {
		name   string
		filter models.NewsFilter
		want   int
	}{
		{"category is case insensitive", models.NewsFilter{Category: str("politics")}, 1},
		{"query matches title", models.NewsFilter{Query: "cup final"}, 1},
		{"query matches summary and title", models.NewsFilter{Query: "budget"}, 2},
		{"query is case insensitive", models.NewsFilter{Query: "RUTO"}, 2},
		{"percent sign is literal", models.NewsFilter{Query: "100%"}, 1},
		{"percent is not a wildcard", models.NewsFilter{Query: "%"}, 1},
		{"query and category combine", models.NewsFilter{Query: "budget", Category: str("Business")}, 1},
		{"no match", models.NewsFilter{Query: "zebra"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.filter.Limit = 100
			_, total, err := repo.ListArticles(ctx, tt.filter)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			// The test database holds only the rows this file inserts, so the count is exact.
			if total != tt.want {
				t.Fatalf("total = %d, want %d", total, tt.want)
			}
		})
	}
}

func TestListArticlesPublisherAndPaging(t *testing.T) {
	pool := testPool(t)
	repo := NewNewsRepo(pool)
	ctx := context.Background()

	var sourceID string
	err := pool.QueryRow(ctx, `
		INSERT INTO news_sources (name, url, type, outlet, publisher, active)
		VALUES ('test-source', 'https://example.com', 'rss', 'other', 'Test Publisher', false)
		ON CONFLICT (name) DO UPDATE SET publisher = EXCLUDED.publisher
		RETURNING id::text`).Scan(&sourceID)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	sameTime := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var articles []models.NewsArticle
	for i := 0; i < 5; i++ {
		published := sameTime
		articles = append(articles, models.NewsArticle{
			URL:         "test://page-" + string(rune('a'+i)),
			Title:       "Paging article",
			PublishedAt: &published,
		})
	}
	seedTestArticles(t, pool, articles...)
	if _, err := pool.Exec(ctx, `UPDATE news_articles SET source_id = $1 WHERE url LIKE 'test://page-%'`, sourceID); err != nil {
		t.Fatalf("attach source: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM news_sources WHERE name = 'test-source'`) })

	first, _, err := repo.ListArticles(ctx, models.NewsFilter{Query: "Paging article", Limit: 2})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	second, _, err := repo.ListArticles(ctx, models.NewsFilter{Query: "Paging article", Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}

	seen := map[string]bool{}
	for _, a := range append(first, second...) {
		if seen[a.ID.String()] {
			t.Fatalf("article %s appeared on two pages; ordering is not stable", a.ID)
		}
		seen[a.ID.String()] = true
		if a.Source == nil || *a.Source != "Test Publisher" {
			t.Fatalf("source = %v, want the publisher name", a.Source)
		}
	}
}

func TestBackfillCategoriesIsIdempotent(t *testing.T) {
	pool := testPool(t)
	repo := NewNewsRepo(pool)
	seedTestArticles(t, pool,
		models.NewsArticle{URL: "test://backfill", Title: "Backfill target", Category: str("NATIONAL NEWS")},
	)
	ctx := context.Background()

	if _, err := repo.BackfillCategories(ctx); err != nil {
		t.Fatalf("first backfill: %v", err)
	}
	var got string
	if err := pool.QueryRow(ctx, `SELECT category FROM news_articles WHERE url = 'test://backfill'`).Scan(&got); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got != "News" {
		t.Fatalf("category = %q, want News", got)
	}

	again, err := repo.BackfillCategories(ctx)
	if err != nil {
		t.Fatalf("second backfill: %v", err)
	}
	if again != 0 {
		t.Fatalf("second backfill changed %d rows, want 0", again)
	}
}
