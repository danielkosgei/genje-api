package scraper

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"jalada/internal/config"
	"jalada/internal/repository"
)

type Scheduler struct {
	fetcher  *RSSFetcher
	newsRepo *repository.NewsRepo
	interval time.Duration
}

func NewScheduler(newsRepo *repository.NewsRepo, cfg config.AggregationConfig) *Scheduler {
	fetcher := NewRSSFetcher(newsRepo, cfg.UserAgent, cfg.RequestTimeout)
	return &Scheduler{
		fetcher:  fetcher,
		newsRepo: newsRepo,
		interval: cfg.Interval,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("starting news scraper scheduler")

	s.run(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("news scraper scheduler stopped")
			return
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

func (s *Scheduler) run(ctx context.Context) {
	log.Debug().Msg("starting news scrape cycle")
	start := time.Now()

	s.fetcher.FetchAll(ctx)
	LinkMentions(ctx, s.newsRepo)

	log.Debug().Dur("duration", time.Since(start)).Msg("news scrape cycle complete")
}
