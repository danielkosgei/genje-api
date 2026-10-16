package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestClaimsReadRatingAndSource(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM claims WHERE source_url LIKE 'test://%'`)
	})

	var politicianID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM politicians LIMIT 1`).Scan(&politicianID); err != nil {
		t.Skipf("no politician to attach a claim to: %v", err)
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO claims (politician_id, claim, rating, rated_on, source_title, source_url)
		VALUES ($1, 'test: a checked claim', 'contested', '2026-09-01', 'test: filing', 'test://claim')`, politicianID)
	if err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	claims, err := NewPoliticianRepo(pool).GetClaims(ctx, politicianID)
	if err != nil {
		t.Fatalf("get claims: %v", err)
	}
	for _, c := range claims {
		if c.SourceURL != "test://claim" {
			continue
		}
		if c.Rating != "contested" || c.RatedOn == nil || *c.RatedOn != "2026-09-01" {
			t.Errorf("rating or date wrong: %+v", c)
		}
		if c.SourceTitle == nil || *c.SourceTitle != "test: filing" {
			t.Errorf("source title wrong: %+v", c.SourceTitle)
		}
		return
	}
	t.Fatal("seeded claim not returned")
}
