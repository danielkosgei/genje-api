package models

// ClaimCounts is how many rated claims a politician has in each rating. Every count is computed from the claims table.
type ClaimCounts struct {
	Confirmed int `json:"confirmed"`
	Unproven  int `json:"unproven"`
	Contested int `json:"contested"`
}

// ProfilePolicyPosition is one position a politician has taken, from a manifesto they published. SourceURL is the document.
type ProfilePolicyPosition struct {
	Sector      string  `json:"sector"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	SourceURL   *string `json:"source_url"`
}

// PoliticianProfile holds the figures the dossier and comparison screens show beside the record.
// A field is null when nothing has been sourced for it yet; it is never estimated.
type PoliticianProfile struct {
	Slug            string                  `json:"slug"`
	Age             *int                    `json:"age"`
	CurrentPost     *string                 `json:"current_post"`
	YearsInOffice   *int                    `json:"years_in_office"`
	HomeCounty      *string                 `json:"home_county"`
	ClaimCounts     ClaimCounts             `json:"claim_counts"`
	PolicyPositions []ProfilePolicyPosition `json:"policy_positions"`
	Links           []PoliticianLink        `json:"links"`
}

// PartyProfile holds the figures for a party page. Counts come from memberships and candidacies.
type PartyProfile struct {
	Slug             string  `json:"slug"`
	Members          int     `json:"members"`
	Aspirants        int     `json:"aspirants"`
	CountiesFielding int     `json:"counties_fielding"`
	LeaderSlug       *string `json:"leader_slug"`
	LeaderName       *string `json:"leader_name"`
}

// SearchHit is one match in a search group.
type SearchHit struct {
	Kind     string  `json:"kind"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	Subtitle *string `json:"subtitle"`
}

// SearchResults groups matches by what they are, so a client can show each group under its own heading.
type SearchResults struct {
	Politicians []SearchHit `json:"politicians"`
	Parties     []SearchHit `json:"parties"`
	Counties    []SearchHit `json:"counties"`
	Coalitions  []SearchHit `json:"coalitions"`
	Elections   []SearchHit `json:"elections"`
}

// RaceCandidate is one candidate in a county race, with the party they stand for.
type RaceCandidate struct {
	PoliticianSlug string  `json:"politician_slug"`
	Name           string  `json:"name"`
	Party          *string `json:"party"`
	Status         string  `json:"status"`
}

// CountyRace is the governor race in one county. Candidates is empty until candidacies are sourced.
type CountyRace struct {
	CountyCode string          `json:"county_code"`
	Candidates []RaceCandidate `json:"candidates"`
}

// PollCandidate is one candidate's average share across the polls in a summary.
type PollCandidate struct {
	CandidateName string  `json:"candidate_name"`
	Party         *string `json:"party"`
	AverageShare  float64 `json:"average_share"`
}

// PollSummary averages the most recent poll from each pollster for one race, over a fixed window.
// Leader and Margin are null when there are fewer than two candidates with results.
type PollSummary struct {
	Race       string          `json:"race"`
	WindowDays int             `json:"window_days"`
	PollCount  int             `json:"poll_count"`
	Leader     *PollCandidate  `json:"leader"`
	Margin     *float64        `json:"margin"`
	Candidates []PollCandidate `json:"candidates"`
}

// PoliticianLink is a cited link about a politician: a profile page or their own website.
// PublisherType is "official" or "third_party". CheckedOn is when the link was last confirmed to exist.
type PoliticianLink struct {
	Kind          string  `json:"kind"`
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Publisher     string  `json:"publisher"`
	PublisherType string  `json:"publisher_type"`
	CheckedOn     string  `json:"checked_on"`
	CheckNote     *string `json:"check_note"`
}
