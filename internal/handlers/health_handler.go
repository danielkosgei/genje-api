package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool    *pgxpool.Pool
	startAt time.Time
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool, startAt: time.Now()}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbOK := true

	if err := h.pool.Ping(ctx); err != nil {
		dbOK = false
	}

	status := "healthy"
	code := http.StatusOK
	if !dbOK {
		status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, map[string]interface{}{
		"status":   status,
		"database": dbOK,
		"uptime":   time.Since(h.startAt).String(),
	})
}

func (h *HealthHandler) Home(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiSchema())
}

func apiSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Jalada",
		"version":     "1.0.0",
		"description": "A dossier on Kenya's 2027 elections.",
		"base_url":    "/v1",
		"endpoints":   endpointList(),
		"schemas":     schemaDefinitions(),
	}
}

func endpointList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"path":        "/health",
			"method":      "GET",
			"description": "Service health check",
			"response":    "HealthResponse",
		},
		// --- Politicians ---
		{
			"path":        "/v1/politicians",
			"method":      "GET",
			"description": "List all politicians with optional search and filtering",
			"parameters": []map[string]interface{}{
				{"name": "q", "in": "query", "type": "string", "description": "Search by name"},
				{"name": "party_id", "in": "query", "type": "uuid", "description": "Filter by party UUID"},
				{"name": "county_id", "in": "query", "type": "uuid", "description": "Filter by county UUID"},
				{"name": "position", "in": "query", "type": "string", "description": "Filter by position type"},
				{"name": "limit", "in": "query", "type": "integer", "default": 20, "description": "Results per page"},
				{"name": "offset", "in": "query", "type": "integer", "default": 0, "description": "Pagination offset"},
			},
			"response": "PaginatedResponse<PoliticianSummary>",
		},
		{
			"path":        "/v1/politicians/{slug}",
			"method":      "GET",
			"description": "Full politician dossier  - bio, education, career, party history, candidacies, integrity flags",
			"parameters": []map[string]interface{}{
				{"name": "slug", "in": "path", "type": "string", "required": true, "description": "Politician URL slug (e.g. william-ruto)"},
			},
			"response": "PoliticianDossier",
		},
		{
			"path":        "/v1/politicians/{slug}/news",
			"method":      "GET",
			"description": "News articles mentioning this politician",
			"parameters": []map[string]interface{}{
				{"name": "slug", "in": "path", "type": "string", "required": true},
				{"name": "limit", "in": "query", "type": "integer", "default": 20},
				{"name": "offset", "in": "query", "type": "integer", "default": 0},
			},
			"response": "PaginatedResponse<NewsArticle>",
		},
		{
			"path":        "/v1/politicians/{slug}/voting-record",
			"method":      "GET",
			"description": "Parliamentary voting record from Hansard",
			"response":    "VotingRecord[]",
		},
		{
			"path":        "/v1/politicians/{slug}/court-cases",
			"method":      "GET",
			"description": "Court cases and legal proceedings",
			"response":    "CourtCase[]",
		},
		{
			"path":        "/v1/politicians/{slug}/promises",
			"method":      "GET",
			"description": "Campaign promises and fulfilment status",
			"response":    "Promise[]",
		},
		{
			"path":        "/v1/politicians/{slug}/achievements",
			"method":      "GET",
			"description": "Notable achievements (tagged, sourced, timestamped)",
			"response":    "Achievement[]",
		},
		{
			"path":        "/v1/politicians/{slug}/controversies",
			"method":      "GET",
			"description": "Controversies and scandals (tagged, sourced, timestamped)",
			"response":    "Controversy[]",
		},
		{
			"path":        "/v1/politicians/{slug}/affiliations",
			"method":      "GET",
			"description": "Political affiliations graph  - allies, family in politics, business connections",
			"response":    "Affiliation[]",
		},
		{
			"path":        "/v1/politicians/{slug}/assets",
			"method":      "GET",
			"description": "Declared assets from EACC filings",
			"response":    "AssetDeclaration[]",
		},
		{
			"path":        "/v1/politicians/{slug}/attendance",
			"method":      "GET",
			"description": "Parliamentary attendance record",
			"response":    "AttendanceRecord[]",
		},
		{
			"path":        "/v1/politicians/{slug}/sentiment",
			"method":      "GET",
			"description": "Public sentiment analysis for this politician",
			"response":    "SentimentSnapshot[]",
		},
		{
			"path":        "/v1/politicians/{slug}/events",
			"method":      "GET",
			"description": "Events and rallies associated with this politician",
			"response":    "Event[]",
		},
		// --- Parties ---
		{
			"path":        "/v1/parties",
			"method":      "GET",
			"description": "List all political parties",
			"response":    "Party[]",
		},
		{
			"path":        "/v1/parties/{slug}",
			"method":      "GET",
			"description": "Party detail with current members list",
			"parameters": []map[string]interface{}{
				{"name": "slug", "in": "path", "type": "string", "required": true, "description": "Party URL slug (e.g. uda, odm)"},
			},
			"response": "PartyDetail",
		},
		// --- Coalitions ---
		{
			"path":        "/v1/coalitions",
			"method":      "GET",
			"description": "List all political coalitions",
			"response":    "Coalition[]",
		},
		{
			"path":        "/v1/coalitions/{slug}",
			"method":      "GET",
			"description": "Coalition detail with member parties",
			"response":    "CoalitionDetail",
		},
		// --- Elections ---
		{
			"path":        "/v1/elections",
			"method":      "GET",
			"description": "List all elections (completed and upcoming)",
			"response":    "Election[]",
		},
		{
			"path":        "/v1/elections/{id}",
			"method":      "GET",
			"description": "Election detail",
			"response":    "Election",
		},
		{
			"path":        "/v1/elections/{id}/candidates",
			"method":      "GET",
			"description": "Candidates registered for this election",
			"response":    "Candidacy[]",
		},
		{
			"path":        "/v1/elections/{id}/results",
			"method":      "GET",
			"description": "Election results by constituency",
			"response":    "ElectionResult[]",
		},
		{
			"path":        "/v1/elections/{id}/timeline",
			"method":      "GET",
			"description": "Election timeline milestones (nominations, campaigns, voting, results)",
			"response":    "TimelineMilestone[]",
		},
		// --- Geography ---
		{
			"path":        "/v1/counties",
			"method":      "GET",
			"description": "List all 47 Kenya counties",
			"response":    "County[]",
		},
		{
			"path":        "/v1/counties/{code}",
			"method":      "GET",
			"description": "County detail by county code",
			"response":    "County",
		},
		{
			"path":        "/v1/counties/{code}/constituencies",
			"method":      "GET",
			"description": "List constituencies within a county",
			"response":    "Constituency[]",
		},
		{
			"path":        "/v1/constituencies/{code}",
			"method":      "GET",
			"description": "Constituency detail by code",
			"response":    "Constituency",
		},
		{
			"path":        "/v1/constituencies/{code}/candidates",
			"method":      "GET",
			"description": "Candidates vying in this constituency",
			"response":    "Candidacy[]",
		},
		{
			"path":        "/v1/constituencies/{code}/wards",
			"method":      "GET",
			"description": "Wards within this constituency",
			"response":    "Ward[]",
		},
		{
			"path":        "/v1/constituencies/{code}/polling-stations",
			"method":      "GET",
			"description": "IEBC polling stations in this constituency",
			"response":    "PollingStation[]",
		},
		// --- News ---
		{
			"path":        "/v1/news",
			"method":      "GET",
			"description": "Aggregated news articles from Kenyan media (auto-updated every 15 minutes)",
			"parameters": []map[string]interface{}{
				{"name": "election_related", "in": "query", "type": "boolean", "description": "Filter to election-related articles only"},
				{"name": "source_id", "in": "query", "type": "uuid", "description": "Filter by news source"},
				{"name": "limit", "in": "query", "type": "integer", "default": 20},
				{"name": "offset", "in": "query", "type": "integer", "default": 0},
			},
			"response": "PaginatedResponse<NewsArticle>",
		},
		{
			"path":        "/v1/news/{id}",
			"method":      "GET",
			"description": "Single news article detail",
			"response":    "NewsArticle",
		},
		{
			"path":        "/v1/sources",
			"method":      "GET",
			"description": "Official data sources used by Jalada (IEBC, EACC, Hansard, etc.)",
			"response":    "Source[]",
		},
		// --- Analytics ---
		{
			"path":        "/v1/analytics/trending",
			"method":      "GET",
			"description": "Trending politicians ranked by recent news mentions",
			"response":    "TrendingItem[]",
		},
		{
			"path":        "/v1/analytics/sentiment",
			"method":      "GET",
			"description": "Aggregate sentiment analysis across candidates",
			"response":    "SentimentSnapshot[]",
		},
		{
			"path":        "/v1/analytics/promises",
			"method":      "GET",
			"description": "Promise fulfilment stats across all politicians",
			"response":    "PromiseStats",
		},
		{
			"path":        "/v1/analytics/integrity",
			"method":      "GET",
			"description": "Integrity flags summary (EACC cases, court cases, Chapter 6 compliance)",
			"response":    "IntegrityStats",
		},
		{
			"path":        "/v1/analytics/attendance",
			"method":      "GET",
			"description": "Parliamentary attendance rankings",
			"response":    "AttendanceStats",
		},
		// --- Timeline & Events ---
		{
			"path":        "/v1/timeline",
			"method":      "GET",
			"description": "2027 election timeline milestones",
			"response":    "TimelineMilestone[]",
		},
		{
			"path":        "/v1/events",
			"method":      "GET",
			"description": "Political events  - rallies, public appearances, parliamentary sessions",
			"parameters": []map[string]interface{}{
				{"name": "type", "in": "query", "type": "string", "description": "Filter by event type (rally, parliamentary, public_appearance)"},
				{"name": "limit", "in": "query", "type": "integer", "default": 20},
				{"name": "offset", "in": "query", "type": "integer", "default": 0},
			},
			"response": "PaginatedResponse<Event>",
		},
	}
}

func schemaDefinitions() map[string]interface{} {
	return map[string]interface{}{
		"PoliticianSummary": map[string]interface{}{
			"description": "Brief politician listing used in search results and party member lists",
			"fields": map[string]string{
				"id":            "uuid",
				"slug":          "string  - URL-safe identifier",
				"first_name":    "string",
				"last_name":     "string",
				"status":        "string  - active | deceased | retired | inactive",
				"photo_url":     "string | null",
				"current_party": "string | null  - party name",
			},
		},
		"PoliticianDossier": map[string]interface{}{
			"description": "Complete politician profile with all available data",
			"fields": map[string]string{
				"id":             "uuid",
				"slug":           "string",
				"first_name":     "string",
				"last_name":      "string",
				"other_names":    "string | null",
				"date_of_birth":  "datetime | null",
				"date_of_death":  "datetime | null  - set when status is deceased",
				"gender":         "string  - male | female | other",
				"status":         "string  - active | deceased | retired | inactive",
				"bio":            "string | null",
				"photo_url":      "string | null",
				"education":      "array  - [{institution, degree, year}]",
				"career_history": "array  - [{role, period}]",
				"current_party":  "PartyMembership | null",
				"party_history":  "PartyMembership[]",
				"candidacies":    "CandidacyDetail[]",
				"integrity_flags": "IntegrityFlag[]",
				"created_at":     "datetime",
				"updated_at":     "datetime",
			},
		},
		"Party": map[string]interface{}{
			"description": "Political party",
			"fields": map[string]string{
				"id":           "uuid",
				"name":         "string",
				"abbreviation": "string",
				"slug":         "string",
				"founded_date": "datetime",
				"ideology":     "string | null",
				"website":      "string | null",
				"logo_url":     "string | null",
				"status":       "string  - active | dissolved | merged",
				"created_at":   "datetime",
				"updated_at":   "datetime",
			},
		},
		"PartyDetail": map[string]interface{}{
			"description": "Party with its current elected members",
			"fields": map[string]string{
				"party":   "Party",
				"members": "PoliticianSummary[]",
			},
		},
		"PartyMembership": map[string]interface{}{
			"description": "A politician's membership in a party",
			"fields": map[string]string{
				"id":            "uuid",
				"politician_id": "uuid",
				"party_id":      "uuid",
				"party_name":    "string",
				"role":          "string | null  - e.g. Party Leader, Secretary General",
				"joined_date":   "datetime",
				"left_date":     "datetime | null",
				"created_at":    "datetime",
			},
		},
		"Election": map[string]interface{}{
			"description": "A general or by-election",
			"fields": map[string]string{
				"id":            "uuid",
				"name":          "string",
				"election_date": "datetime",
				"type":          "string  - general | by_election | referendum",
				"status":        "string  - upcoming | ongoing | completed",
				"created_at":    "datetime",
				"updated_at":    "datetime",
			},
		},
		"TimelineMilestone": map[string]interface{}{
			"description": "An election timeline milestone",
			"fields": map[string]string{
				"id":             "uuid",
				"election_id":    "uuid",
				"title":          "string",
				"description":    "string | null",
				"milestone_type": "string  - nomination_start | nomination_end | registration_open | registration_close | campaign_start | silence_period | election_day | results_announcement | inauguration",
				"date":           "datetime",
				"status":         "string  - upcoming | current | completed",
			},
		},
		"County": map[string]interface{}{
			"description": "One of Kenya's 47 counties",
			"fields": map[string]string{
				"id":   "uuid",
				"code": "string  - 3-digit county code",
				"name": "string",
				"slug": "string",
			},
		},
		"Constituency": map[string]interface{}{
			"description": "A parliamentary constituency within a county",
			"fields": map[string]string{
				"id":                "uuid",
				"county_id":         "uuid",
				"code":              "string  - 3-digit constituency code",
				"name":              "string",
				"slug":              "string",
				"registered_voters": "integer",
			},
		},
		"NewsArticle": map[string]interface{}{
			"description": "A news article scraped from Kenyan media RSS feeds",
			"fields": map[string]string{
				"id":                   "uuid",
				"source_id":            "uuid",
				"title":                "string",
				"content":              "string",
				"summary":              "string | null",
				"url":                  "string  - original article URL",
				"author":               "string | null",
				"published_at":         "datetime",
				"scraped_at":           "datetime",
				"is_election_related":  "boolean",
				"created_at":           "datetime",
			},
		},
		"TrendingItem": map[string]interface{}{
			"description": "A trending politician ranked by recent news mentions",
			"fields": map[string]string{
				"id":    "uuid",
				"name":  "string",
				"slug":  "string",
				"score": "integer  - number of recent mentions",
				"type":  "string  - politician",
			},
		},
		"PaginatedResponse": map[string]interface{}{
			"description": "Wrapper for paginated list endpoints",
			"fields": map[string]string{
				"data":     "array  - the results for this page",
				"total":    "integer  - total matching records",
				"limit":    "integer  - page size used",
				"offset":   "integer  - current offset",
				"has_more": "boolean  - true if more pages exist",
			},
		},
		"HealthResponse": map[string]interface{}{
			"description": "Service health status",
			"fields": map[string]string{
				"status":   "string  - healthy | unhealthy",
				"database": "boolean  - database connectivity",
				"uptime":   "string  - server uptime duration",
			},
		},
		"ErrorResponse": map[string]interface{}{
			"description": "Standard error response returned on 4xx/5xx",
			"fields": map[string]string{
				"error":   "string  - HTTP status text",
				"code":    "integer  - HTTP status code",
				"message": "string  - human-readable error detail",
			},
		},
	}
}
