package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"jalada/internal/config"
	"jalada/internal/database"
	"jalada/internal/handlers"
	"jalada/internal/repository"
	"jalada/internal/scraper"
	"jalada/internal/seeder"
	"jalada/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	setupLogger(cfg)

	log.Info().Str("env", cfg.Server.Env).Msg("starting Jalada")

	if err := database.RunMigrations(cfg.Database.URL); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := seeder.Seed(ctx, pool); err != nil {
		log.Fatal().Err(err).Msg("failed to seed data")
	}

	// Repositories
	politicianRepo := repository.NewPoliticianRepo(pool)
	partyRepo := repository.NewPartyRepo(pool)
	electionRepo := repository.NewElectionRepo(pool)
	geographyRepo := repository.NewGeographyRepo(pool)
	newsRepo := repository.NewNewsRepo(pool)
	eventRepo := repository.NewEventRepo(pool)
	sentimentRepo := repository.NewSentimentRepo(pool)
	analyticsRepo := repository.NewAnalyticsRepo(pool)

	// Services
	politicianSvc := services.NewPoliticianService(politicianRepo, newsRepo, sentimentRepo, eventRepo)
	electionSvc := services.NewElectionService(electionRepo)
	timelineSvc := services.NewTimelineService(eventRepo)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo, sentimentRepo)

	// Handlers
	h := &handlers.Handlers{
		Health:     handlers.NewHealthHandler(pool),
		Politician: handlers.NewPoliticianHandler(politicianSvc),
		Party:      handlers.NewPartyHandler(partyRepo),
		Election:   handlers.NewElectionHandler(electionSvc),
		News:       handlers.NewNewsHandler(newsRepo),
		Geography:  handlers.NewGeographyHandler(geographyRepo),
		Analytics:  handlers.NewAnalyticsHandler(analyticsSvc),
		Timeline:   handlers.NewTimelineHandler(timelineSvc),
	}

	router := handlers.NewRouter(h)

	newsScheduler := scraper.NewScheduler(newsRepo, cfg.Aggregation)
	go newsScheduler.Start(ctx)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}

func setupLogger(cfg *config.Config) {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	if !cfg.Log.JSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
}
