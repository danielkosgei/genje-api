package repository

import (
	"context"
	"testing"
)

// seedTestPoll inserts one poll under a test: prefix and removes it when the test ends.
func seedTestPoll(t *testing.T) {
	t.Helper()
	pool := testPool(t)
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM polls WHERE pollster_id IN (SELECT id FROM pollsters WHERE name LIKE 'test:%')`)
		_, _ = pool.Exec(ctx, `DELETE FROM pollsters WHERE name LIKE 'test:%'`)
	})
	_, err := pool.Exec(ctx, `
		WITH ps AS (
			INSERT INTO pollsters (name) VALUES ('test: Example Pollster') RETURNING id
		), p AS (
			INSERT INTO polls (pollster_id, race, sample_size, margin_of_error, published_on, source_url)
			SELECT id, 'president', 1200, 2.5, '2026-09-30', 'test://poll' FROM ps RETURNING id
		)
		INSERT INTO poll_results (poll_id, candidate_name, party, share)
		SELECT id, 'Candidate A', 'Party A', 48.5 FROM p
		UNION ALL
		SELECT id, 'Candidate B', 'Party B', 41.0 FROM p`)
	if err != nil {
		t.Fatalf("seed poll: %v", err)
	}
}

func TestPollListReadsResultsAndNumbers(t *testing.T) {
	seedTestPoll(t)
	pool := testPool(t)
	repo := NewPollRepo(pool)

	polls, _, err := repo.List(context.Background(), "president", 100, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found bool
	for _, p := range polls {
		if p.Pollster != "test: Example Pollster" {
			continue
		}
		found = true
		if p.SourceURL != "test://poll" || p.PublishedOn != "2026-09-30" {
			t.Errorf("source or date wrong: %+v", p)
		}
		if p.SampleSize == nil || *p.SampleSize != 1200 {
			t.Errorf("sample size wrong: %+v", p.SampleSize)
		}
		if p.MarginOfError == nil || *p.MarginOfError != 2.5 {
			t.Errorf("margin wrong: %+v", p.MarginOfError)
		}
		if len(p.Results) != 2 || p.Results[0].Share != 48.5 || p.Results[1].CandidateName != "Candidate B" {
			t.Errorf("results wrong or unordered: %+v", p.Results)
		}
	}
	if !found {
		t.Fatal("seeded poll not returned")
	}
}
