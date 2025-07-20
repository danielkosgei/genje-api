package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"genje-api/internal/config"
	"genje-api/internal/database"
	"genje-api/internal/handlers"
	"genje-api/internal/middleware"
	"genje-api/internal/repository"
	"genje-api/internal/services"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	sourceRepo := repository.NewSourceRepository(db)
	engagementRepo := repository.NewEngagementRepository(db)

	// Initialize services
	aggregatorService := services.NewAggregatorService(articleRepo, sourceRepo, cfg.Aggregator)
	summarizerService := services.NewSummarizerService(articleRepo)
	trendingService := services.NewTrendingService(db, articleRepo, engagementRepo, summarizerService)

	// Seed initial news sources
	if err := sourceRepo.SeedInitialSources(); err != nil {
		log.Printf("Warning: Failed to seed news sources: %v", err)
	}

	// Start background aggregation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go aggregatorService.StartBackgroundAggregation(ctx)

	// Start background trending cache updates
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // Update every 30 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Update trending cache for all time windows
				timeWindows := []string{"1h", "6h", "12h", "24h", "7d"}
				for _, window := range timeWindows {
					// Cache trending scores using the trending service
					_, err := trendingService.GetAdvancedTrendingArticles(100, window)
					if err != nil {
						log.Printf("Warning: Failed to update trending cache for %s: %v", window, err)
					}
				}
			}
		}
	}()

	// Initialize handlers
	h := handlers.New(articleRepo, sourceRepo, engagementRepo, aggregatorService, summarizerService, trendingService)

	// Setup router
	r := setupRouter(h)

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Root endpoint - API info
	r.Get("/", h.GetAPIInfo)

	// Health check
	r.Get("/health", h.Health)

	// API routes
	r.Route("/v1", func(r chi.Router) {
		// API metadata
		r.Get("/openapi.json", h.GetOpenAPISpec)
		r.Get("/schema", h.GetAPISchema)
		r.Get("/status", h.GetSystemStatus)

		// Articles resource
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", h.GetArticles)                   // GET /v1/articles
			r.Get("/{id}", h.GetArticle)                // GET /v1/articles/123
			r.Get("/{id}/summary", h.SummarizeArticle)  // GET /v1/articles/123/summary
			r.Post("/{id}/summary", h.SummarizeArticle) // POST /v1/articles/123/summary

			// Engagement tracking
			r.Post("/{id}/engage", h.TrackEngagement)         // POST /v1/articles/123/engage
			r.Get("/{id}/engagement", h.GetArticleEngagement) // GET /v1/articles/123/engagement

			// Article collections and filters
			r.Get("/search", h.SearchArticles)                         // GET /v1/articles/search?q=term
			r.Get("/feed", h.GetArticlesFeed)                          // GET /v1/articles/feed (cursor pagination)
			r.Get("/trending", h.GetTrendingArticles)                  // GET /v1/articles/trending (uses advanced algorithm)
			r.Get("/trending/advanced", h.GetAdvancedTrendingArticles) // GET /v1/articles/trending/advanced
			r.Get("/top-engaged", h.GetTopEngagedArticles)             // GET /v1/articles/top-engaged
			r.Get("/recent", h.GetRecentArticles)                      // GET /v1/articles/recent
		})

		// Sources resource
		r.Route("/sources", func(r chi.Router) {
			r.Get("/", h.GetSources)                 // GET /v1/sources
			r.Post("/", h.CreateSource)              // POST /v1/sources
			r.Get("/{id}", h.GetSource)              // GET /v1/sources/123
			r.Put("/{id}", h.UpdateSource)           // PUT /v1/sources/123
			r.Patch("/{id}", h.UpdateSource)         // PATCH /v1/sources/123
			r.Delete("/{id}", h.DeleteSource)        // DELETE /v1/sources/123
			r.Post("/{id}/refresh", h.RefreshSource) // POST /v1/sources/123/refresh

			// Source sub-resources
			r.Get("/{id}/articles", h.GetArticlesBySource) // GET /v1/sources/123/articles
		})

		// Categories resource
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", h.GetCategories)                        // GET /v1/categories
			r.Get("/{name}/articles", h.GetArticlesByCategory) // GET /v1/categories/sports/articles
		})

		// Statistics resource
		r.Route("/stats", func(r chi.Router) {
			r.Get("/", h.GetGlobalStats)               // GET /v1/stats
			r.Get("/sources", h.GetSourceStats)        // GET /v1/stats/sources
			r.Get("/categories", h.GetCategoryStats)   // GET /v1/stats/categories
			r.Get("/timeline", h.GetTimelineStats)     // GET /v1/stats/timeline
			r.Get("/engagement", h.GetEngagementStats) // GET /v1/stats/engagement
		})

		// Engagement resource
		r.Route("/engagement", func(r chi.Router) {
			r.Post("/cache/refresh", h.RefreshTrendingCache) // POST /v1/engagement/cache/refresh
		})

		// Authority resource
		r.Route("/authority", func(r chi.Router) {
			r.Get("/sources/{name}", h.GetSourceAuthority)     // GET /v1/authority/sources/Standard%20Media
			r.Post("/sources/{name}", h.UpdateSourceAuthority) // POST /v1/authority/sources/Standard%20Media
		})

		// Trends resource
		r.Get("/trends", h.GetTrends) // GET /v1/trends

		// System operations
		r.Post("/refresh", h.RefreshNews) // POST /v1/refresh
	})

	return r
}
