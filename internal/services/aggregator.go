package services

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"genje-api/internal/config"
	"genje-api/internal/models"
	"genje-api/internal/repository"

	"github.com/mmcdole/gofeed"
)

type AggregatorService struct {
	articleRepo *repository.ArticleRepository
	sourceRepo  *repository.SourceRepository
	config      config.AggregatorConfig
	client      *http.Client
	parser      *gofeed.Parser
}

func NewAggregatorService(articleRepo *repository.ArticleRepository, sourceRepo *repository.SourceRepository, cfg config.AggregatorConfig) *AggregatorService {
	// Create HTTP client with better settings for Docker environments
	client := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	parser := gofeed.NewParser()
	parser.Client = client

	return &AggregatorService{
		articleRepo: articleRepo,
		sourceRepo:  sourceRepo,
		config:      cfg,
		client:      client,
		parser:      parser,
	}
}

func (s *AggregatorService) StartBackgroundAggregation(ctx context.Context) {
	// Run immediately
	if err := s.AggregateNews(ctx); err != nil {
		log.Printf("ERROR: Initial aggregation failed: %v", err)
	}

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Background aggregation stopped")
			return
		case <-ticker.C:
			if err := s.AggregateNews(ctx); err != nil {
				log.Printf("ERROR: Scheduled aggregation failed: %v", err)
			}
		}
	}
}

func (s *AggregatorService) AggregateNews(ctx context.Context) error {
	log.Printf("=== Starting news aggregation at %s ===", time.Now().Format("2006-01-02 15:04:05"))

	sources, err := s.sourceRepo.GetSourcesForAggregation()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	log.Printf("Found %d active news sources to process", len(sources))

	totalProcessed := 0
	successCount := 0
	errorCount := 0

	for i, source := range sources {
		select {
		case <-ctx.Done():
			log.Println("Aggregation cancelled")
			return ctx.Err()
		default:
		}

		log.Printf("--- Processing source %d/%d: %s ---", i+1, len(sources), source.Name)

		processed, err := s.processFeed(ctx, source)
		if err != nil {
			log.Printf("ERROR: Failed to process feed %s: %v", source.Name, err)
			errorCount++
			continue
		}

		totalProcessed += processed
		successCount++

		// Small delay between requests to be respectful
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("=== News aggregation completed at %s ===", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("SUMMARY: Processed %d articles from %d sources (%d successful, %d failed)",
		totalProcessed, len(sources), successCount, errorCount)

	return nil
}

func (s *AggregatorService) processFeed(ctx context.Context, source models.NewsSource) (int, error) {
	log.Printf("Fetching from %s (%s)...", source.Name, source.FeedURL)

	req, err := http.NewRequestWithContext(ctx, "GET", source.FeedURL, nil)
	if err != nil {
		log.Printf("ERROR: Failed to create request for %s: %v", source.Name, err)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to avoid 403 errors and improve compatibility
	req.Header.Set("User-Agent", s.config.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, application/atom+xml, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	log.Printf("Making request to %s with User-Agent: %s", source.FeedURL, s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("ERROR: Failed to fetch %s: %v", source.Name, err)
		return 0, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Response from %s: Status=%d, Content-Type=%s, Content-Length=%s",
		source.Name, resp.StatusCode, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"))

	if resp.StatusCode != 200 {
		log.Printf("ERROR: HTTP error from %s: %d %s", source.Name, resp.StatusCode, resp.Status)
		return 0, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Handle gzip decompression if needed
	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Printf("ERROR: Failed to create gzip reader for %s: %v", source.Name, err)
			return 0, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read the body content for debugging
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("ERROR: Failed to read response body from %s: %v", source.Name, err)
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if content looks like XML
	if !strings.Contains(string(bodyBytes), "<") {
		log.Printf("ERROR: Content from %s does not appear to be XML", source.Name)
		return 0, fmt.Errorf("content is not XML")
	}

	// Parse the feed from the bytes
	feed, err := s.parser.ParseString(string(bodyBytes))
	if err != nil {
		log.Printf("ERROR: Failed to parse feed from %s: %v", source.Name, err)
		return 0, fmt.Errorf("failed to parse feed: %w", err)
	}

	log.Printf("Parsed feed from %s: Found %d items", source.Name, len(feed.Items))

	articles := s.convertFeedItems(feed.Items, source)
	if len(articles) == 0 {
		log.Printf("WARNING: No valid articles found from %s after processing %d feed items", source.Name, len(feed.Items))
		return 0, nil
	}

	log.Printf("Converted %d feed items to %d valid articles from %s", len(feed.Items), len(articles), source.Name)

	if err := s.articleRepo.CreateArticlesBatch(articles); err != nil {
		log.Printf("ERROR: Failed to save articles from %s: %v", source.Name, err)
		return 0, fmt.Errorf("failed to save articles: %w", err)
	}

	log.Printf("SUCCESS: Processed %d articles from %s", len(articles), source.Name)
	return len(articles), nil
}

func (s *AggregatorService) convertFeedItems(items []*gofeed.Item, source models.NewsSource) []models.Article {
	var articles []models.Article
	skipped := 0

	for i, item := range items {
		if item.Title == "" || item.Link == "" {
			log.Printf("SKIP: Item %d from %s missing title or link (Title='%s', Link='%s')",
				i+1, source.Name, item.Title, item.Link)
			skipped++
			continue
		}

		article := models.Article{
			Title:    strings.TrimSpace(item.Title),
			URL:      strings.TrimSpace(item.Link),
			Source:   source.Name,
			Category: source.Category,
		}

		// Set published date
		if item.PublishedParsed != nil {
			article.PublishedAt = *item.PublishedParsed
		} else if item.Published != "" {
			// Try to parse the published string manually
			if parsed, err := time.Parse(time.RFC1123Z, item.Published); err == nil {
				article.PublishedAt = parsed
			} else if parsed, err := time.Parse(time.RFC1123, item.Published); err == nil {
				article.PublishedAt = parsed
			} else {
				log.Printf("DEBUG: Could not parse published date '%s' for item from %s, using current time",
					item.Published, source.Name)
				article.PublishedAt = time.Now()
			}
		} else {
			article.PublishedAt = time.Now()
		}

		// Set author
		if item.Author != nil && item.Author.Name != "" {
			article.Author = strings.TrimSpace(item.Author.Name)
		}

		// Set content
		if item.Content != "" {
			article.Content = item.Content
		} else if item.Description != "" {
			article.Content = item.Description
		} else {
			log.Printf("DEBUG: No content found for item '%s' from %s", article.Title, source.Name)
		}

		// Clean and limit content
		article.Content = s.cleanContent(article.Content)

		// Set image URL
		if item.Image != nil && item.Image.URL != "" {
			article.ImageURL = strings.TrimSpace(item.Image.URL)
		}

		// Additional validation
		if len(article.Title) > 500 {
			log.Printf("SKIP: Title too long (%d chars) for item from %s", len(article.Title), source.Name)
			skipped++
			continue
		}

		articles = append(articles, article)
	}

	if skipped > 0 {
		log.Printf("INFO: Skipped %d invalid items from %s, processed %d valid items",
			skipped, source.Name, len(articles))
	}

	return articles
}

func (s *AggregatorService) cleanContent(content string) string {
	content = strings.TrimSpace(content)

	// Limit content size
	if len(content) > s.config.MaxContentSize {
		content = content[:s.config.MaxContentSize] + "..."
	}

	return content
}
