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

	// Initialize services
	aggregatorService := services.NewAggregatorService(articleRepo, sourceRepo, cfg.Aggregator)
	summarizerService := services.NewSummarizerService(articleRepo)

	// Seed initial news sources
	if err := sourceRepo.SeedInitialSources(); err != nil {
		log.Printf("Warning: Failed to seed news sources: %v", err)
	}

	// Start background aggregation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go aggregatorService.StartBackgroundAggregation(ctx)

	// Initialize handlers
	h := handlers.New(articleRepo, sourceRepo, aggregatorService, summarizerService)

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
			r.Get("/", h.GetArticles)           // GET /v1/articles
			r.Get("/{id}", h.GetArticle)        // GET /v1/articles/123
			r.Post("/{id}/summary", h.SummarizeArticle) // POST /v1/articles/123/summary
			
			// Article collections and filters
			r.Get("/search", h.SearchArticles)   // GET /v1/articles/search?q=term
			r.Get("/feed", h.GetArticlesFeed)    // GET /v1/articles/feed (cursor pagination)
			r.Get("/trending", h.GetTrendingArticles) // GET /v1/articles/trending
			r.Get("/recent", h.GetRecentArticles)     // GET /v1/articles/recent
		})
		
		// Sources resource
		r.Route("/sources", func(r chi.Router) {
			r.Get("/", h.GetSources)            // GET /v1/sources
			r.Post("/", h.CreateSource)         // POST /v1/sources
			r.Get("/{id}", h.GetSource)         // GET /v1/sources/123
			r.Put("/{id}", h.UpdateSource)      // PUT /v1/sources/123
			r.Patch("/{id}", h.UpdateSource)    // PATCH /v1/sources/123
			r.Delete("/{id}", h.DeleteSource)   // DELETE /v1/sources/123
			r.Post("/{id}/refresh", h.RefreshSource) // POST /v1/sources/123/refresh
			
			// Source sub-resources
			r.Get("/{id}/articles", h.GetArticlesBySource) // GET /v1/sources/123/articles
		})
		
		// Categories resource
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", h.GetCategories)         // GET /v1/categories
			r.Get("/{name}/articles", h.GetArticlesByCategory) // GET /v1/categories/sports/articles
		})
		
		// Statistics resource
		r.Route("/stats", func(r chi.Router) {
			r.Get("/", h.GetGlobalStats)        // GET /v1/stats
			r.Get("/sources", h.GetSourceStats) // GET /v1/stats/sources
			r.Get("/categories", h.GetCategoryStats) // GET /v1/stats/categories
			r.Get("/timeline", h.GetTimelineStats)   // GET /v1/stats/timeline
		})
		
		// Trends resource
		r.Get("/trends", h.GetTrends)           // GET /v1/trends
		
		// System operations
		r.Post("/refresh", h.RefreshNews)       // POST /v1/refresh
	})

	return r
} 