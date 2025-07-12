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

	// Health check
	r.Get("/health", h.Health)

	// API routes
	r.Route("/v1", func(r chi.Router) {
		r.Get("/articles", h.GetArticles)
		r.Get("/articles/{id}", h.GetArticle)
		r.Post("/articles/{id}/summarize", h.SummarizeArticle)
		r.Get("/sources", h.GetSources)
		r.Get("/categories", h.GetCategories)
		r.Post("/refresh", h.RefreshNews)
	})

	return r
} 