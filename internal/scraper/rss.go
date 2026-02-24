package scraper

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"

	"jalada/internal/models"
	"jalada/internal/repository"
)

var electionKeywords = []string{
	"election", "iebc", "ballot", "campaign", "manifesto", "coalition",
	"aspirant", "candidate", "vote", "poll", "gubernatorial", "senator",
	"mp ", "mca ", "president", "deputy president", "governor",
	"constituency", "county assembly", "parliament", "nomination",
	"party primaries", "running mate", "political party", "tallying",
	"electoral", "voter registration", "civic education",
	"kenya kwanza", "azimio", "uda", "odm",
}

type RSSFetcher struct {
	newsRepo  *repository.NewsRepo
	parser    *gofeed.Parser
	client    *http.Client
	userAgent string
}

func NewRSSFetcher(newsRepo *repository.NewsRepo, userAgent string, timeout time.Duration) *RSSFetcher {
	return &RSSFetcher{
		newsRepo:  newsRepo,
		parser:    gofeed.NewParser(),
		client:    &http.Client{Timeout: timeout},
		userAgent: userAgent,
	}
}

func (f *RSSFetcher) FetchAll(ctx context.Context) {
	sources, err := f.newsRepo.GetActiveSources(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get active news sources")
		return
	}

	var totalNew, totalSkipped int
	for _, source := range sources {
		if source.FeedURL == nil || *source.FeedURL == "" {
			continue
		}
		newCount, skipCount := f.fetchSource(ctx, source)
		totalNew += newCount
		totalSkipped += skipCount
	}
	log.Info().Int("new", totalNew).Int("skipped", totalSkipped).Int("sources", len(sources)).Msg("RSS fetch cycle complete")
}

func (f *RSSFetcher) fetchSource(ctx context.Context, source models.NewsSource) (newCount, skipCount int) {
	req, err := http.NewRequestWithContext(ctx, "GET", *source.FeedURL, nil)
	if err != nil {
		log.Warn().Err(err).Str("source", source.Name).Msg("failed to create request")
		return 0, 0
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		log.Warn().Err(err).Str("source", source.Name).Msg("failed to fetch feed")
		return 0, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Str("source", source.Name).Msg("non-200 response from feed")
		return 0, 0
	}

	feed, err := f.parser.Parse(resp.Body)
	if err != nil {
		log.Warn().Err(err).Str("source", source.Name).Msg("failed to parse feed")
		return 0, 0
	}

	for _, item := range feed.Items {
		if item.Link == "" {
			continue
		}

		exists, err := f.newsRepo.ArticleExistsByURL(ctx, item.Link)
		if err != nil {
			log.Warn().Err(err).Str("url", item.Link).Msg("failed to check article existence")
			continue
		}
		if exists {
			skipCount++
			continue
		}

		article := f.itemToArticle(item, &source)
		if err := f.newsRepo.InsertArticle(ctx, article); err != nil {
			log.Warn().Err(err).Str("url", item.Link).Msg("failed to insert article")
			continue
		}
		newCount++
	}

	log.Debug().Str("source", source.Name).Int("new", newCount).Int("skipped", skipCount).Msg("fetched feed")
	return newCount, skipCount
}

func (f *RSSFetcher) itemToArticle(item *gofeed.Item, source *models.NewsSource) *models.NewsArticle {
	article := &models.NewsArticle{
		SourceID:          &source.ID,
		Title:             item.Title,
		URL:               item.Link,
		IsElectionRelated: isElectionRelated(item.Title, item.Description),
	}

	if item.Description != "" {
		article.Summary = &item.Description
	}
	if item.Content != "" {
		article.Content = &item.Content
	} else if item.Description != "" {
		article.Content = &item.Description
	}
	if item.Author != nil {
		article.Author = &item.Author.Name
	}
	if len(item.Authors) > 0 && article.Author == nil {
		article.Author = &item.Authors[0].Name
	}
	if imgURL := extractImageURL(item); imgURL != "" {
		article.ImageURL = &imgURL
	}
	if item.PublishedParsed != nil {
		article.PublishedAt = item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		article.PublishedAt = item.UpdatedParsed
	}
	if len(item.Categories) > 0 {
		cat := item.Categories[0]
		article.Category = &cat
	}

	return article
}

func extractImageURL(item *gofeed.Item) string {
	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}

	for _, enc := range item.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") && enc.URL != "" {
			return enc.URL
		}
	}

	for _, ns := range []string{"media", "Media"} {
		if exts, ok := item.Extensions[ns]; ok {
			for _, tag := range []string{"content", "thumbnail"} {
				if elems, ok := exts[tag]; ok && len(elems) > 0 {
					if url := elems[0].Attrs["url"]; url != "" {
						return url
					}
				}
			}
		}
	}

	return ""
}

func isElectionRelated(title, description string) bool {
	text := strings.ToLower(title + " " + description)
	for _, kw := range electionKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
