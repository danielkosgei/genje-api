package services

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"genje-api/internal/models"
	"genje-api/internal/repository"
)

type TrendingService struct {
	db                *sql.DB
	articleRepo       *repository.ArticleRepository
	engagementRepo    *repository.EngagementRepository
	summarizerService *SummarizerService
}

func NewTrendingService(db *sql.DB, articleRepo *repository.ArticleRepository,
	engagementRepo *repository.EngagementRepository, summarizerService *SummarizerService) *TrendingService {
	return &TrendingService{
		db:                db,
		articleRepo:       articleRepo,
		engagementRepo:    engagementRepo,
		summarizerService: summarizerService,
	}
}

// GetAdvancedTrendingArticles implements the sophisticated 5-factor trending algorithm
func (s *TrendingService) GetAdvancedTrendingArticles(limit int, timeWindow string) ([]models.EnhancedTrendingArticle, error) {
	// First, try to get from cache
	cached, err := s.getCachedTrending(limit, timeWindow)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Calculate fresh trending scores
	trending, err := s.calculateTrendingScores(limit, timeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate trending scores: %w", err)
	}

	// Cache the results
	go s.cacheTrendingResults(trending, timeWindow)

	return trending, nil
}

// calculateTrendingScores implements the 5-factor algorithm
func (s *TrendingService) calculateTrendingScores(limit int, timeWindow string) ([]models.EnhancedTrendingArticle, error) {
	hours := s.parseTimeWindow(timeWindow)

	// Get articles from the time window
	query := `
		SELECT a.id, a.title, COALESCE(a.content, ''), COALESCE(a.summary, ''), a.url, 
			COALESCE(a.author, ''), a.source, a.published_at, a.created_at, 
			COALESCE(a.category, 'general'), COALESCE(a.image_url, '')
		FROM articles a
		WHERE a.published_at >= DATETIME('now', '-' || ? || ' hours')
		ORDER BY a.published_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, hours, limit*3) // Get more to filter and rank
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary,
			&article.URL, &article.Author, &article.Source, &article.PublishedAt,
			&article.CreatedAt, &article.Category, &article.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	// Calculate trending scores for each article
	var trendingArticles []models.EnhancedTrendingArticle
	for _, article := range articles {
		enhanced, err := s.calculateArticleTrendingScore(article, timeWindow)
		if err != nil {
			continue // Skip articles with calculation errors
		}
		trendingArticles = append(trendingArticles, *enhanced)
	}

	// Sort by trending score and limit results
	s.sortByTrendingScore(trendingArticles)
	if len(trendingArticles) > limit {
		trendingArticles = trendingArticles[:limit]
	}

	return trendingArticles, nil
}

// calculateArticleTrendingScore implements the 5-factor scoring algorithm
func (s *TrendingService) calculateArticleTrendingScore(article models.Article, timeWindow string) (*models.EnhancedTrendingArticle, error) {
	// Factor 1: Engagement-Based Scoring
	engagement, err := s.engagementRepo.GetArticleEngagement(article.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get engagement: %w", err)
	}
	engagementScore := s.calculateEngagementScore(*engagement)

	// Factor 2: Velocity-Based Trending
	velocity, err := s.engagementRepo.GetEngagementVelocity(article.ID, timeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to get velocity: %w", err)
	}
	velocityScore := s.normalizeVelocity(velocity)

	// Factor 3: Source Authority Weighting
	sourceAuthority, err := s.engagementRepo.GetSourceAuthority(article.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to get source authority: %w", err)
	}
	authorityScore := s.calculateAuthorityScore(*sourceAuthority)

	// Factor 4: Content Analysis (NLP-based)
	contentScore := s.calculateContentScore(article)

	// Factor 5: Time Decay Function
	recencyScore := s.calculateRecencyScore(article.PublishedAt, timeWindow)

	// Combine all factors with weights
	weights := models.ScoreBreakdown{
		EngagementWeight: 0.35, // 35% - Most important
		VelocityWeight:   0.25, // 25% - Trending momentum
		AuthorityWeight:  0.20, // 20% - Source credibility
		ContentWeight:    0.10, // 10% - Content quality
		RecencyWeight:    0.10, // 10% - Time relevance
	}

	finalScore := (engagementScore * weights.EngagementWeight) +
		(velocityScore * weights.VelocityWeight) +
		(authorityScore * weights.AuthorityWeight) +
		(contentScore * weights.ContentWeight) +
		(recencyScore * weights.RecencyWeight)

	// Determine trending reason
	trendingReason := s.determineTrendingReason(engagementScore, velocityScore, authorityScore, contentScore, recencyScore)

	return &models.EnhancedTrendingArticle{
		Article:            article,
		Engagement:         *engagement,
		TrendingScore:      finalScore,
		EngagementVelocity: velocity,
		RecencyScore:       recencyScore,
		AuthorityScore:     authorityScore,
		ContentScore:       contentScore,
		TrendingReason:     trendingReason,
		ScoreBreakdown:     weights,
	}, nil
}

// Factor 1: Engagement-Based Scoring
func (s *TrendingService) calculateEngagementScore(engagement models.ArticleEngagement) float64 {
	// Weighted engagement score
	score := float64(engagement.Views)*0.1 +
		float64(engagement.Shares)*0.4 +
		float64(engagement.Comments)*0.3 +
		float64(engagement.Likes)*0.2

	// Normalize to 0-1 scale (assuming max realistic engagement)
	maxExpectedEngagement := 10000.0
	return math.Min(score/maxExpectedEngagement, 1.0)
}

// Factor 2: Velocity-Based Trending (already implemented in engagement repo)
func (s *TrendingService) normalizeVelocity(velocity float64) float64 {
	// Velocity is already normalized between -1 and 1, convert to 0-1
	return (velocity + 1.0) / 2.0
}

// Factor 3: Source Authority Weighting
func (s *TrendingService) calculateAuthorityScore(authority models.SourceAuthority) float64 {
	// Combine authority, credibility, and reach scores
	return (authority.AuthorityScore*0.4 + authority.CredibilityScore*0.3 + authority.ReachScore*0.3)
}

// Factor 4: Content Analysis (NLP-based)
func (s *TrendingService) calculateContentScore(article models.Article) float64 {
	score := 0.0

	// Title analysis
	titleScore := s.analyzeTitleQuality(article.Title)
	score += titleScore * 0.4

	// Content length and quality
	contentScore := s.analyzeContentQuality(article.Content)
	score += contentScore * 0.3

	// Keyword trending analysis
	keywordScore := s.analyzeKeywordTrending(article.Title + " " + article.Content)
	score += keywordScore * 0.3

	return math.Min(score, 1.0)
}

// Factor 5: Time Decay Function
func (s *TrendingService) calculateRecencyScore(publishedAt time.Time, timeWindow string) float64 {
	now := time.Now()
	age := now.Sub(publishedAt)

	// Different decay rates for different time windows
	var halfLife time.Duration
	switch timeWindow {
	case "1h":
		halfLife = 30 * time.Minute
	case "6h":
		halfLife = 2 * time.Hour
	case "12h":
		halfLife = 4 * time.Hour
	case "24h":
		halfLife = 8 * time.Hour
	case "7d":
		halfLife = 2 * 24 * time.Hour
	default:
		halfLife = 8 * time.Hour
	}

	// Exponential decay: score = e^(-age/halfLife)
	decay := math.Exp(-age.Seconds() / halfLife.Seconds())
	return math.Max(decay, 0.01) // Minimum score to avoid zero
}

// Content analysis helper functions
func (s *TrendingService) analyzeTitleQuality(title string) float64 {
	score := 0.5 // Base score

	// Length optimization (ideal 50-60 characters)
	titleLen := len(title)
	if titleLen >= 40 && titleLen <= 70 {
		score += 0.2
	}

	// Presence of numbers (often indicates data/facts)
	if matched, _ := regexp.MatchString(`\d+`, title); matched {
		score += 0.1
	}

	// Question format (engaging)
	if strings.Contains(title, "?") {
		score += 0.1
	}

	// Breaking news indicators
	breakingWords := []string{"breaking", "urgent", "alert", "developing", "exclusive"}
	titleLower := strings.ToLower(title)
	for _, word := range breakingWords {
		if strings.Contains(titleLower, word) {
			score += 0.1
			break
		}
	}

	return math.Min(score, 1.0)
}

func (s *TrendingService) analyzeContentQuality(content string) float64 {
	if len(content) == 0 {
		return 0.3 // Low score for missing content
	}

	score := 0.5 // Base score

	// Content length (ideal range)
	contentLen := len(content)
	if contentLen >= 500 && contentLen <= 3000 {
		score += 0.3
	} else if contentLen >= 200 {
		score += 0.1
	}

	// Paragraph structure (multiple paragraphs indicate better formatting)
	paragraphs := strings.Count(content, "\n\n") + strings.Count(content, "</p>")
	if paragraphs >= 3 {
		score += 0.2
	}

	return math.Min(score, 1.0)
}

func (s *TrendingService) analyzeKeywordTrending(text string) float64 {
	// Simplified keyword analysis - in production, you'd use more sophisticated NLP
	trendingKeywords := []string{
		"kenya", "nairobi", "president", "government", "economy", "election",
		"covid", "health", "education", "technology", "business", "sports",
		"breaking", "exclusive", "developing", "urgent", "alert",
	}

	textLower := strings.ToLower(text)
	matches := 0
	for _, keyword := range trendingKeywords {
		if strings.Contains(textLower, keyword) {
			matches++
		}
	}

	// Normalize based on keyword density
	return math.Min(float64(matches)/10.0, 1.0)
}

// Helper functions
func (s *TrendingService) parseTimeWindow(timeWindow string) int {
	switch timeWindow {
	case "1h":
		return 1
	case "6h":
		return 6
	case "12h":
		return 12
	case "24h":
		return 24
	case "7d":
		return 24 * 7
	default:
		return 24
	}
}

func (s *TrendingService) determineTrendingReason(engagement, velocity, authority, content, recency float64) string {
	maxScore := math.Max(math.Max(math.Max(engagement, velocity), math.Max(authority, content)), recency)

	switch {
	case maxScore == velocity && velocity > 0.7:
		return "Rapidly gaining engagement"
	case maxScore == engagement && engagement > 0.6:
		return "High user engagement"
	case maxScore == authority && authority > 0.8:
		return "Authoritative source"
	case maxScore == content && content > 0.7:
		return "High-quality content"
	case maxScore == recency && recency > 0.8:
		return "Breaking news"
	default:
		return "Trending across multiple factors"
	}
}

func (s *TrendingService) sortByTrendingScore(articles []models.EnhancedTrendingArticle) {
	for i := 0; i < len(articles)-1; i++ {
		for j := i + 1; j < len(articles); j++ {
			if articles[i].TrendingScore < articles[j].TrendingScore {
				articles[i], articles[j] = articles[j], articles[i]
			}
		}
	}
}

// Caching functions for performance
func (s *TrendingService) getCachedTrending(limit int, timeWindow string) ([]models.EnhancedTrendingArticle, error) {
	// Check if cache is fresh (within 15 minutes)
	query := `
		SELECT tc.article_id, tc.trending_score, tc.engagement_velocity, tc.recency_score,
			tc.authority_score, tc.content_score, tc.calculated_at
		FROM trending_cache tc
		WHERE tc.time_window = ? AND tc.calculated_at >= DATETIME('now', '-15 minutes')
		ORDER BY tc.trending_score DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, timeWindow, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cached []models.EnhancedTrendingArticle
	for rows.Next() {
		var articleID int
		var trendingScore, velocity, recency, authority, content float64
		var calculatedAt time.Time

		err := rows.Scan(&articleID, &trendingScore, &velocity, &recency, &authority, &content, &calculatedAt)
		if err != nil {
			continue
		}

		// Get full article data
		article, err := s.articleRepo.GetArticleByID(articleID)
		if err != nil {
			continue
		}

		engagement, _ := s.engagementRepo.GetArticleEngagement(articleID)
		if engagement == nil {
			engagement = &models.ArticleEngagement{ArticleID: articleID}
		}

		enhanced := models.EnhancedTrendingArticle{
			Article:            *article,
			Engagement:         *engagement,
			TrendingScore:      trendingScore,
			EngagementVelocity: velocity,
			RecencyScore:       recency,
			AuthorityScore:     authority,
			ContentScore:       content,
			TrendingReason:     s.determineTrendingReason(0, velocity, authority, content, recency),
		}

		cached = append(cached, enhanced)
	}

	return cached, nil
}

func (s *TrendingService) cacheTrendingResults(articles []models.EnhancedTrendingArticle, timeWindow string) {
	// Clear old cache for this time window
	if _, err := s.db.Exec("DELETE FROM trending_cache WHERE time_window = ?", timeWindow); err != nil {
		// Log error but continue
		return
	}

	// Insert new cache entries
	for _, article := range articles {
		if _, err := s.db.Exec(`
			INSERT INTO trending_cache (article_id, time_window, trending_score, engagement_velocity,
				recency_score, authority_score, content_score)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, article.ID, timeWindow, article.TrendingScore, article.EngagementVelocity,
			article.RecencyScore, article.AuthorityScore, article.ContentScore); err != nil {
			// Log error but continue with next article
			continue
		}
	}
}
