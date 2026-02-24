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
	PhotoURL      *string          `json:"photo_url"`
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
	if count >= 25 {
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
	if count >= 400 {
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
			`INSERT INTO politicians (slug, first_name, last_name, other_names, date_of_birth, date_of_death, gender, status, bio, photo_url, education, career_history)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			 ON CONFLICT (slug) DO UPDATE SET photo_url = EXCLUDED.photo_url`,
			p.Slug, p.FirstName, p.LastName, p.OtherNames, p.DateOfBirth, p.DateOfDeath, p.Gender, statusOrDefault(p.Status), p.Bio,
			p.PhotoURL, educationJSON(p.Education), careerJSON(p.CareerHistory),
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

	type milestone struct {
		title         string
		description   string
		milestoneType string
		date          string
		status        string
	}

	milestones := []milestone{
		{
			"Continuous Voter Registration Begins",
			"IEBC resumed CVR on September 29, 2025 across all 290 constituency offices and 57 Huduma Centres per Article 88(4) of the Constitution. Target: 6.3 million new voters to reach 28.5 million total. Source: IEBC gazette notice, The Star, Capital FM.",
			"registration_open", "2025-09-29", "passed",
		},
		{
			"Election Technology Deployment Target",
			"IEBC targets having election technology systems (KIEMS kits, BVRS replacements, results transmission) fully deployed by June 2026. Source: IEBC 2027 roadmap, Citizen Digital.",
			"other", "2026-06-30", "upcoming",
		},
		{
			"Mass Voter Registration Opens",
			"30-day enhanced mass voter registration exercise across all polling stations. Eligible citizens 18+ with valid national ID or passport. Source: IEBC 2027 roadmap, Citizen Digital.",
			"other", "2027-03-29", "upcoming",
		},
		{
			"Recruitment of Temporary Election Officials Begins",
			"IEBC begins recruiting presiding officers, deputy presiding officers, and polling clerks for the general election. Source: IEBC 2027 roadmap.",
			"other", "2027-04-01", "upcoming",
		},
		{
			"Mass Voter Registration Closes",
			"Final day for new voter registration before the 2027 general election. Source: IEBC 2027 roadmap, Citizen Digital.",
			"registration_close", "2027-04-29", "upcoming",
		},
		{
			"Procurement of Election Materials Begins",
			"IEBC begins procurement of ballot papers, tamper-evident envelopes, indelible ink, and other election materials. Source: IEBC 2027 roadmap.",
			"other", "2027-05-01", "upcoming",
		},
		{
			"Political Party Primaries Deadline",
			"Deadline for political parties to conclude internal primaries, resolve disputes, and submit final candidate lists to IEBC. Estimated from IEBC 2027 roadmap (parties conclude by May 2027) and 2022 precedent.",
			"nomination_end", "2027-05-15", "upcoming",
		},
		{
			"IEBC Candidate Nomination Opens",
			"IEBC opens nomination and registration of candidates for all elective positions. Presidential candidates must present supporter lists with at least 2,000 voters from a majority of counties. Estimated from 2022 precedent (May 29 - June 6, 2022). Source: Elections Act.",
			"nomination_start", "2027-05-29", "upcoming",
		},
		{
			"IEBC Candidate Nomination Closes",
			"Final day for candidates to submit nomination papers to IEBC. Followed by a 10-day dispute resolution window. Estimated from 2022 precedent. Source: Elections Act, IEBC guidelines.",
			"other", "2027-06-07", "upcoming",
		},
		{
			"Official Campaign Period Begins",
			"Campaigns permitted daily from 7:00 AM to 6:00 PM. Campaigns must stop 48 hours before polling day per Elections Act. Start date is approximately 45 days before election (2022 precedent: May 29). Source: Elections Act Cap. 7.",
			"campaign_start", "2027-06-26", "upcoming",
		},
		{
			"Campaign Silence Period Begins",
			"All campaign activities must cease 48 hours before polling day per the Elections Act. No campaign rallies, advertisements, or voter canvassing permitted from this date. Source: Elections Act Section 13.",
			"silence_period", "2027-08-08", "upcoming",
		},
		{
			"Election Day",
			"Constitutionally fixed date for the 2027 Kenya General Election per Article 101(1) of the Constitution. Voters elect President, Governors, Senators, Members of National Assembly, Woman Representatives, and Members of County Assembly. Polling hours: 6:00 AM to 5:00 PM. Source: Constitution of Kenya, Article 101.",
			"election_day", "2027-08-10", "upcoming",
		},
		{
			"Presidential Results Declaration Deadline",
			"IEBC Chairperson must declare presidential results within 7 days of the election per Article 138(2) of the Constitution. Other results (Governor, Senator, MP, Woman Rep, MCA) declared at constituency/county level. Source: Constitution of Kenya, Article 138.",
			"results_announcement", "2027-08-17", "upcoming",
		},
		{
			"Election Petition Filing Deadline",
			"Any petition challenging presidential election results must be filed at the Supreme Court within 7 days of the results declaration per Article 140(1). The Supreme Court has 14 days to hear and determine the petition. Source: Constitution of Kenya, Article 140.",
			"other", "2027-08-24", "upcoming",
		},
		{
			"Inauguration of President-Elect",
			"The President-elect is sworn in on the first Tuesday following the 14th day after results declaration (if no petition) per Article 141 of the Constitution. Date estimated assuming results on August 17 and no petition. Source: Constitution of Kenya, Article 141.",
			"inauguration", "2027-09-07", "upcoming",
		},
	}

	var electionID string
	err := pool.QueryRow(ctx, `SELECT id FROM elections WHERE name = $1`, "2027 Kenya General Election").Scan(&electionID)
	if err != nil {
		return fmt.Errorf("find 2027 election: %w", err)
	}

	_, _ = pool.Exec(ctx, `DELETE FROM election_timeline WHERE election_id = $1`, electionID)

	for _, m := range milestones {
		_, _ = pool.Exec(ctx,
			`INSERT INTO election_timeline (election_id, title, description, milestone_type, date, status)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			electionID, m.title, m.description, m.milestoneType, m.date, m.status,
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
