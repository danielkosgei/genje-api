package scraper

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type politicianMatcher struct {
	entries []matchEntry
}

type matchEntry struct {
	ID       uuid.UUID
	Keywords []string
}

func newPoliticianMatcher(politicians []models.PoliticianSummary) *politicianMatcher {
	entries := make([]matchEntry, 0, len(politicians))
	for _, p := range politicians {
		keywords := []string{
			strings.ToLower(p.FirstName + " " + p.LastName),
			strings.ToLower(p.LastName),
		}
		entries = append(entries, matchEntry{ID: p.ID, Keywords: keywords})
	}
	return &politicianMatcher{entries: entries}
}

func (m *politicianMatcher) FindMentions(text string) []uuid.UUID {
	lower := strings.ToLower(text)
	var matched []uuid.UUID
	seen := make(map[uuid.UUID]bool)

	for _, e := range m.entries {
		if seen[e.ID] {
			continue
		}
		for _, kw := range e.Keywords {
			if len(kw) >= 4 && strings.Contains(lower, kw) {
				matched = append(matched, e.ID)
				seen[e.ID] = true
				break
			}
		}
	}
	return matched
}

func LinkMentions(ctx context.Context, newsRepo *repository.NewsRepo) {
	politicians, err := newsRepo.GetAllPoliticianNames(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to load politician names for matching")
		return
	}

	matcher := newPoliticianMatcher(politicians)

	filter := models.NewsFilter{Limit: 100, Offset: 0}
	articles, _, err := newsRepo.ListArticles(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("failed to list recent articles for mention linking")
		return
	}

	var totalMentions int
	for _, article := range articles {
		text := article.Title
		if article.Summary != nil {
			text += " " + *article.Summary
		}
		if article.Content != nil {
			text += " " + *article.Content
		}

		matches := matcher.FindMentions(text)
		for _, politicianID := range matches {
			if err := newsRepo.InsertMention(ctx, article.URL, politicianID); err != nil {
				log.Warn().Err(err).Str("url", article.URL).Msg("failed to insert mention")
			} else {
				totalMentions++
			}
		}
	}

	if totalMentions > 0 {
		log.Info().Int("mentions", totalMentions).Int("articles", len(articles)).Msg("politician mentions linked")
	}
}
