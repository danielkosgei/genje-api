package seeder

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

//go:embed data/*.json
var dataFS embed.FS

type countyData struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type constituencyData struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	CountyCode string `json:"county_code"`
}

type partyData struct {
	Name         string  `json:"name"`
	Abbreviation string  `json:"abbreviation"`
	Slug         string  `json:"slug"`
	Ideology     *string `json:"ideology"`
	Website      *string `json:"website"`
	Status       string  `json:"status"`
	FoundedDate  *string `json:"founded_date"`
}

type educationEntry struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Year        int    `json:"year,omitempty"`
}

type careerEntry struct {
	Role   string `json:"role"`
	Period string `json:"period"`
}

type politicianData struct {
	FirstName     string           `json:"first_name"`
	LastName      string           `json:"last_name"`
	OtherNames    *string          `json:"other_names"`
	Slug          string           `json:"slug"`
	DateOfBirth   *string          `json:"date_of_birth"`
	DateOfDeath   *string          `json:"date_of_death"`
	Gender        string           `json:"gender"`
	Status        string           `json:"status"`
	Bio           *string          `json:"bio"`
	Education     []educationEntry `json:"education"`
	CareerHistory []careerEntry    `json:"career_history"`
	PartySlug     *string          `json:"party_slug"`
}

type newsSourceData struct {
	Name    string  `json:"name"`
	URL     string  `json:"url"`
	FeedURL *string `json:"feed_url"`
	Type    string  `json:"type"`
	Outlet  *string `json:"outlet"`
	Active  bool    `json:"active"`
}

func loadJSON[T any](filename string) ([]T, error) {
	data, err := dataFS.ReadFile("data/" + filename)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	return items, nil
}

func tableCount(ctx context.Context, pool *pgxpool.Pool, table string) (int, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
	return count, err
}

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	log.Info().Msg("checking seed data...")

	if err := seedCounties(ctx, pool); err != nil {
		return fmt.Errorf("seed counties: %w", err)
	}
	if err := seedConstituencies(ctx, pool); err != nil {
		return fmt.Errorf("seed constituencies: %w", err)
	}
	if err := seedParties(ctx, pool); err != nil {
		return fmt.Errorf("seed parties: %w", err)
	}
	if err := seedPoliticians(ctx, pool); err != nil {
		return fmt.Errorf("seed politicians: %w", err)
	}
	if err := seedCoalitions(ctx, pool); err != nil {
		return fmt.Errorf("seed coalitions: %w", err)
	}
	if err := seedElections(ctx, pool); err != nil {
		return fmt.Errorf("seed elections: %w", err)
	}
	if err := seedNewsSources(ctx, pool); err != nil {
		return fmt.Errorf("seed news sources: %w", err)
	}
	if err := seedDataSources(ctx, pool); err != nil {
		return fmt.Errorf("seed data sources: %w", err)
	}

	log.Info().Msg("seed data check complete")
	return nil
}

func seedCounties(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "counties")
	if count >= 47 {
		log.Debug().Int("count", count).Msg("counties already seeded")
		return nil
	}

	counties, err := loadJSON[countyData]("counties.json")
	if err != nil {
		return err
	}

	for _, c := range counties {
		_, err := pool.Exec(ctx,
			`INSERT INTO counties (code, name, slug) VALUES ($1, $2, $3) ON CONFLICT (code) DO NOTHING`,
			c.Code, c.Name, c.Slug,
		)
		if err != nil {
			return fmt.Errorf("insert county %s: %w", c.Name, err)
		}
	}
	log.Info().Int("count", len(counties)).Msg("seeded counties")
	return nil
}

func seedConstituencies(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "constituencies")
	if count >= 290 {
		log.Debug().Int("count", count).Msg("constituencies already seeded")
		return nil
	}

	constituencies, err := loadJSON[constituencyData]("constituencies.json")
	if err != nil {
		return err
	}

	for _, c := range constituencies {
		_, err := pool.Exec(ctx,
			`INSERT INTO constituencies (code, name, slug, county_id)
			 VALUES ($1, $2, $3, (SELECT id FROM counties WHERE code = $4))
			 ON CONFLICT (code) DO NOTHING`,
			c.Code, c.Name, c.Slug, c.CountyCode,
		)
		if err != nil {
			return fmt.Errorf("insert constituency %s: %w", c.Name, err)
		}
	}
	log.Info().Int("count", len(constituencies)).Msg("seeded constituencies")
	return nil
}

func seedParties(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "political_parties")
	if count >= 15 {
		log.Debug().Int("count", count).Msg("parties already seeded")
		return nil
	}

	parties, err := loadJSON[partyData]("parties.json")
	if err != nil {
		return err
	}

	for _, p := range parties {
		_, err := pool.Exec(ctx,
			`INSERT INTO political_parties (name, abbreviation, slug, ideology, website, status, founded_date)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (slug) DO NOTHING`,
			p.Name, p.Abbreviation, p.Slug, p.Ideology, p.Website, p.Status, p.FoundedDate,
		)
		if err != nil {
			return fmt.Errorf("insert party %s: %w", p.Name, err)
		}
	}
	log.Info().Int("count", len(parties)).Msg("seeded parties")
	return nil
}

func seedPoliticians(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "politicians")
	if count >= 50 {
		log.Debug().Int("count", count).Msg("politicians already seeded")
		return nil
	}

	politicians, err := loadJSON[politicianData]("politicians.json")
	if err != nil {
		return err
	}

	statusOrDefault := func(s string) string {
		if s == "" {
			return "active"
		}
		return s
	}
	educationJSON := func(entries []educationEntry) []byte {
		if entries == nil {
			entries = []educationEntry{}
		}
		b, _ := json.Marshal(entries)
		return b
	}
	careerJSON := func(entries []careerEntry) []byte {
		if entries == nil {
			entries = []careerEntry{}
		}
		b, _ := json.Marshal(entries)
		return b
	}

	for _, p := range politicians {
		_, err := pool.Exec(ctx,
			`INSERT INTO politicians (slug, first_name, last_name, other_names, date_of_birth, date_of_death, gender, status, bio, education, career_history)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			 ON CONFLICT (slug) DO NOTHING`,
			p.Slug, p.FirstName, p.LastName, p.OtherNames, p.DateOfBirth, p.DateOfDeath, p.Gender, statusOrDefault(p.Status), p.Bio,
			educationJSON(p.Education), careerJSON(p.CareerHistory),
		)
		if err != nil {
			return fmt.Errorf("insert politician %s %s: %w", p.FirstName, p.LastName, err)
		}

		if p.PartySlug != nil {
			_, _ = pool.Exec(ctx,
				`INSERT INTO party_memberships (politician_id, party_id, joined_date)
				 SELECT p.id, pp.id, NOW()
				 FROM politicians p, political_parties pp
				 WHERE p.slug = $1 AND pp.slug = $2
				 AND NOT EXISTS (
				     SELECT 1 FROM party_memberships pm
				     WHERE pm.politician_id = p.id AND pm.party_id = pp.id AND pm.left_date IS NULL
				 )`,
				p.Slug, *p.PartySlug,
			)
		}
	}
	log.Info().Int("count", len(politicians)).Msg("seeded politicians")
	return nil
}

func seedCoalitions(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "coalitions")
	if count >= 2 {
		log.Debug().Int("count", count).Msg("coalitions already seeded")
		return nil
	}

	type coalition struct {
		name           string
		slug           string
		formedDate     string
		principalSlug  string
		memberSlugs    []string
	}

	coalitions := []coalition{
		{
			name:          "Kenya Kwanza Alliance",
			slug:          "kenya-kwanza",
			formedDate:    "2022-03-12",
			principalSlug: "uda",
			memberSlugs:   []string{"uda", "anc", "ford-kenya", "ccm", "mdg", "tsp", "kup"},
		},
		{
			name:          "Azimio la Umoja - One Kenya Coalition",
			slug:          "azimio-la-umoja",
			formedDate:    "2022-03-12",
			principalSlug: "odm",
			memberSlugs:   []string{"odm", "jubilee-party", "wiper", "dap-k", "narc-kenya", "maendeleo-chap-chap", "paa"},
		},
	}

	for _, c := range coalitions {
		_, err := pool.Exec(ctx,
			`INSERT INTO coalitions (name, slug, formed_date, principal_party_id)
			 VALUES ($1, $2, $3, (SELECT id FROM political_parties WHERE slug = $4))
			 ON CONFLICT (slug) DO NOTHING`,
			c.name, c.slug, c.formedDate, c.principalSlug,
		)
		if err != nil {
			return fmt.Errorf("insert coalition %s: %w", c.name, err)
		}

		for _, memberSlug := range c.memberSlugs {
			_, _ = pool.Exec(ctx,
				`INSERT INTO coalition_members (coalition_id, party_id, joined_at)
				 SELECT co.id, pp.id, $3::date
				 FROM coalitions co, political_parties pp
				 WHERE co.slug = $1 AND pp.slug = $2
				 ON CONFLICT (coalition_id, party_id) DO NOTHING`,
				c.slug, memberSlug, c.formedDate,
			)
		}
	}
	log.Info().Int("count", len(coalitions)).Msg("seeded coalitions")
	return nil
}

func seedElections(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "elections")
	if count >= 2 {
		log.Debug().Int("count", count).Msg("elections already seeded")
		return nil
	}

	type electionSeed struct {
		name string
		date string
		typ  string
		stat string
	}

	elections := []electionSeed{
		{"2022 Kenya General Election", "2022-08-09", "general", "completed"},
		{"2027 Kenya General Election", "2027-08-10", "general", "upcoming"},
	}

	for _, e := range elections {
		_, err := pool.Exec(ctx,
			`INSERT INTO elections (name, election_date, type, status)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT DO NOTHING`,
			e.name, e.date, e.typ, e.stat,
		)
		if err != nil {
			return fmt.Errorf("insert election %s: %w", e.name, err)
		}
	}

	milestones := []struct {
		title         string
		milestoneType string
		date          string
		status        string
	}{
		{"Party Primaries Open", "nomination_start", "2027-03-01", "upcoming"},
		{"Party Primaries Close", "nomination_end", "2027-04-30", "upcoming"},
		{"IEBC Candidate Registration Opens", "registration_open", "2027-05-01", "upcoming"},
		{"IEBC Candidate Registration Closes", "registration_close", "2027-06-01", "upcoming"},
		{"Official Campaign Period Begins", "campaign_start", "2027-06-02", "upcoming"},
		{"Campaign Silence Period", "silence_period", "2027-08-08", "upcoming"},
		{"Election Day", "election_day", "2027-08-10", "upcoming"},
		{"Results Announcement", "results_announcement", "2027-08-17", "upcoming"},
		{"Inauguration", "inauguration", "2027-09-14", "upcoming"},
	}

	var electionID string
	err := pool.QueryRow(ctx, `SELECT id FROM elections WHERE name = $1`, "2027 Kenya General Election").Scan(&electionID)
	if err != nil {
		return fmt.Errorf("find 2027 election: %w", err)
	}

	existingMilestones, _ := tableCount(ctx, pool, "election_timeline")
	if existingMilestones >= len(milestones) {
		return nil
	}

	for _, m := range milestones {
		_, _ = pool.Exec(ctx,
			`INSERT INTO election_timeline (election_id, title, milestone_type, date, status)
			 VALUES ($1, $2, $3, $4, $5)`,
			electionID, m.title, m.milestoneType, m.date, m.status,
		)
	}
	log.Info().Int("count", len(milestones)).Msg("seeded 2027 election timeline")
	return nil
}

func seedNewsSources(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "news_sources")
	if count >= 5 {
		log.Debug().Int("count", count).Msg("news sources already seeded")
		return nil
	}

	sources, err := loadJSON[newsSourceData]("news_sources.json")
	if err != nil {
		return err
	}

	for _, s := range sources {
		_, err := pool.Exec(ctx,
			`INSERT INTO news_sources (name, url, feed_url, type, outlet, active)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT DO NOTHING`,
			s.Name, s.URL, s.FeedURL, s.Type, s.Outlet, s.Active,
		)
		if err != nil {
			return fmt.Errorf("insert news source %s: %w", s.Name, err)
		}
	}
	log.Info().Int("count", len(sources)).Msg("seeded news sources")
	return nil
}

func seedDataSources(ctx context.Context, pool *pgxpool.Pool) error {
	count, _ := tableCount(ctx, pool, "sources")
	if count >= 5 {
		log.Debug().Int("count", count).Msg("data sources already seeded")
		return nil
	}

	type src struct {
		name        string
		url         string
		typ         string
		reliability string
	}

	sources := []src{
		{"Independent Electoral and Boundaries Commission", "https://www.iebc.or.ke", "iebc", "official"},
		{"Ethics and Anti-Corruption Commission", "https://www.eacc.go.ke", "eacc", "official"},
		{"Kenya Gazette", "http://www.kenyalaw.org/kenya_gazette/", "gazette", "official"},
		{"Kenya National Assembly Hansard", "http://www.parliament.go.ke/the-national-assembly/hansard", "hansard", "official"},
		{"Kenya Senate Hansard", "http://www.parliament.go.ke/the-senate/hansard", "hansard", "official"},
		{"Judiciary of Kenya", "https://www.judiciary.go.ke", "court", "official"},
		{"Office of the Registrar of Political Parties", "https://www.orpp.or.ke", "other", "official"},
	}

	now := time.Now()
	for _, s := range sources {
		_, _ = pool.Exec(ctx,
			`INSERT INTO sources (name, url, type, reliability, last_accessed_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			s.name, s.url, s.typ, s.reliability, now,
		)
	}
	log.Info().Int("count", len(sources)).Msg("seeded official data sources")
	return nil
}
