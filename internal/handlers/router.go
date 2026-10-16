package handlers

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"jalada/internal/middleware"
)

type Handlers struct {
	Health     *HealthHandler
	Politician *PoliticianHandler
	Party      *PartyHandler
	Election   *ElectionHandler
	News       *NewsHandler
	Nominee    *NomineeHandler
	Poll       *PollHandler
	Search     *SearchHandler
	Race       *RaceHandler
	Geography  *GeographyHandler
	Analytics  *AnalyticsHandler
	Timeline   *TimelineHandler
}

func NewRouter(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(middleware.CORS()))
	r.Use(middleware.RateLimit(100, 200))

	r.Get("/", h.Health.Home)
	r.Get("/health", h.Health.Health)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/", h.Health.Home)

		// Politicians
		r.Route("/politicians", func(r chi.Router) {
			r.Get("/", h.Politician.List)
			r.Route("/{slug}", func(r chi.Router) {
				r.Get("/", h.Politician.GetDossier)
				r.Get("/news", h.Politician.GetNews)
				r.Get("/voting-record", h.Politician.GetVotingRecord)
				r.Get("/court-cases", h.Politician.GetCourtCases)
				r.Get("/profile", h.Politician.GetProfile)
				r.Get("/promises", h.Politician.GetPromises)
				r.Get("/claims", h.Politician.GetClaims)
				r.Get("/achievements", h.Politician.GetAchievements)
				r.Get("/controversies", h.Politician.GetControversies)
				r.Get("/affiliations", h.Politician.GetAffiliations)
				r.Get("/assets", h.Politician.GetAssets)
				r.Get("/attendance", h.Politician.GetAttendance)
				r.Get("/sentiment", h.Politician.GetSentiment)
				r.Get("/events", h.Politician.GetEvents)
			})
		})

		// Parties
		r.Route("/parties", func(r chi.Router) {
			r.Get("/", h.Party.ListParties)
			r.Get("/{slug}", h.Party.GetParty)
			r.Get("/{slug}/profile", h.Party.GetPartyProfile)
		})

		// Coalitions
		r.Route("/coalitions", func(r chi.Router) {
			r.Get("/", h.Party.ListCoalitions)
			r.Get("/{slug}", h.Party.GetCoalition)
		})

		// Elections
		r.Route("/elections", func(r chi.Router) {
			r.Get("/", h.Election.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.Election.Get)
				r.Get("/candidates", h.Election.GetCandidates)
				r.Get("/results", h.Election.GetResults)
				r.Get("/timeline", h.Election.GetTimeline)
			})
		})

		// Polls (sourced opinion polls; only what has been recorded)
		r.Route("/polls", func(r chi.Router) {
			r.Get("/", h.Poll.ListPolls)
			r.Get("/summary", h.Poll.GetSummary)
		})

		// Search across politicians, parties, counties, coalitions and elections
		r.Get("/search", h.Search.Search)

		// Nominee lists (IEBC party nominees)
		r.Route("/nominees", func(r chi.Router) {
			r.Get("/", h.Nominee.ListNominees)
			r.Get("/headings", h.Nominee.ListNomineeHeadings)
		})

		// Geography
		r.Route("/counties", func(r chi.Router) {
			r.Get("/", h.Geography.ListCounties)
			r.Route("/{code}", func(r chi.Router) {
				r.Get("/", h.Geography.GetCounty)
				r.Get("/race", h.Race.GetCountyRace)
				r.Get("/constituencies", h.Geography.GetConstituencies)
			})
		})
		r.Route("/constituencies", func(r chi.Router) {
			r.Route("/{code}", func(r chi.Router) {
				r.Get("/", h.Geography.GetConstituency)
				r.Get("/candidates", h.Geography.GetCandidates)
				r.Get("/wards", h.Geography.GetWards)
				r.Get("/polling-stations", h.Geography.GetPollingStations)
			})
		})

		// News
		r.Route("/news", func(r chi.Router) {
			r.Get("/", h.News.ListArticles)
			r.Get("/{id}", h.News.GetArticle)
		})
		r.Get("/sources", h.News.ListSources)

		// Analytics
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/sentiment", h.Analytics.Sentiment)
			r.Get("/promises", h.Analytics.Promises)
			r.Get("/integrity", h.Analytics.Integrity)
			r.Get("/attendance", h.Analytics.Attendance)
			r.Get("/trending", h.Analytics.Trending)
		})

		// Timeline and events
		r.Get("/timeline", h.Timeline.GetTimeline)
		r.Get("/events", h.Timeline.ListEvents)
	})

	return r
}
