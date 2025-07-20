package services

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"genje-api/internal/models"
)

type EngagementService struct {
	db *sql.DB
}

func NewEngagementService(db *sql.DB) *EngagementService {
	return &EngagementService{db: db}
}

// TrackEngagement records an engagement event and updates counters
func (s *EngagementService) TrackEngagement(articleID int, req models.EngagementRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			// Log rollback error but don't override the main error
			// In production, this would be logged properly
			_ = err // Satisfy linter
		}
	}()

	// Insert engagement event
	_, err = tx.Exec(`
		INSERT INTO engagement_events (article_id, event_type, user_ip, user_agent, metadata)
		VALUES (?, ?, ?, ?, ?)
	`, articleID, req.EventType, req.UserIP, req.UserAgent, req.Metadata)
	if err != nil {
		return fmt.Errorf("failed to insert engagement event: %w", err)
	}

	// Update or create engagement counters
	_, err = tx.Exec(`
		INSERT INTO article_engagement (article_id, views, shares, comments, likes)
		VALUES (?, 
			CASE WHEN ? = 'view' THEN 1 ELSE 0 END,
			CASE WHEN ? = 'share' THEN 1 ELSE 0 END,
			CASE WHEN ? = 'comment' THEN 1 ELSE 0 END,
			CASE WHEN ? = 'like' THEN 1 ELSE 0 END
		)
		ON CONFLICT(article_id) DO UPDATE SET
			views = views + CASE WHEN ? = 'view' THEN 1 ELSE 0 END,
			shares = shares + CASE WHEN ? = 'share' THEN 1 ELSE 0 END,
			comments = comments + CASE WHEN ? = 'comment' THEN 1 ELSE 0 END,
			likes = likes + CASE WHEN ? = 'like' THEN 1 ELSE 0 END,
			last_updated = CURRENT_TIMESTAMP
	`, articleID, req.EventType, req.EventType, req.EventType, req.EventType,
		req.EventType, req.EventType, req.EventType, req.EventType)
	if err != nil {
		return fmt.Errorf("failed to update engagement counters: %w", err)
	}

	return tx.Commit()
}

// GetArticleEngagement returns engagement metrics for an article
func (s *EngagementService) GetArticleEngagement(articleID int) (*models.ArticleEngagement, error) {
	var engagement models.ArticleEngagement
	err := s.db.QueryRow(`
		SELECT id, article_id, views, shares, comments, likes, last_updated
		FROM article_engagement
		WHERE article_id = ?
	`, articleID).Scan(&engagement.ID, &engagement.ArticleID, &engagement.Views,
		&engagement.Shares, &engagement.Comments, &engagement.Likes, &engagement.LastUpdated)

	if err == sql.ErrNoRows {
		// Return zero engagement if not found
		return &models.ArticleEngagement{
			ArticleID:   articleID,
			Views:       0,
			Shares:      0,
			Comments:    0,
			Likes:       0,
			LastUpdated: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article engagement: %w", err)
	}

	return &engagement, nil
}

// CalculateEngagementVelocity calculates how fast engagement is growing
func (s *EngagementService) CalculateEngagementVelocity(articleID int, timeWindow string) (float64, error) {
	hours := s.parseTimeWindow(timeWindow)

	// Get engagement events in the time window
	rows, err := s.db.Query(`
		SELECT event_type, COUNT(*) as count,
			strftime('%H', timestamp) as hour
		FROM engagement_events
		WHERE article_id = ? 
			AND timestamp >= DATETIME('now', '-' || ? || ' hours')
		GROUP BY event_type, hour
		ORDER BY hour DESC
	`, articleID, hours)
	if err != nil {
		return 0, fmt.Errorf("failed to query engagement velocity: %w", err)
	}
	defer rows.Close()

	// Calculate velocity based on recent vs older engagement
	recentEngagement := 0.0
	olderEngagement := 0.0
	halfWindow := hours / 2

	for rows.Next() {
		var eventType string
		var count int
		var hour int

		if err := rows.Scan(&eventType, &count, &hour); err != nil {
			continue
		}

		weight := s.getEventWeight(eventType)
		weightedCount := float64(count) * weight

		// Recent half vs older half
		if hour <= halfWindow {
			recentEngagement += weightedCount
		} else {
			olderEngagement += weightedCount
		}
	}

	// Calculate velocity (recent engagement / older engagement)
	if olderEngagement == 0 {
		if recentEngagement > 0 {
			return 2.0, nil // High velocity for new trending content
		}
		return 0.0, nil
	}

	velocity := recentEngagement / olderEngagement
	return math.Min(velocity, 5.0), nil // Cap at 5x velocity
}

// CalculateRecencyScore calculates score based on article age
func (s *EngagementService) CalculateRecencyScore(publishedAt time.Time) float64 {
	age := time.Since(publishedAt)
	hours := age.Hours()

	// Exponential decay: newer articles score higher
	// Score drops to ~0.1 after 48 hours
	score := math.Exp(-hours / 24.0)
	return math.Max(score, 0.01) // Minimum score
}

// GetSourceAuthority returns authority score for a source
func (s *EngagementService) GetSourceAuthority(sourceName string) (float64, error) {
	var authority models.SourceAuthority
	err := s.db.QueryRow(`
		SELECT authority_score, credibility_score, reach_score
		FROM source_authority
		WHERE source_name = ?
	`, sourceName).Scan(&authority.AuthorityScore, &authority.CredibilityScore, &authority.ReachScore)

	if err == sql.ErrNoRows {
		// Default authority for unknown sources
		return 0.5, nil
	}
	if err != nil {
		return 0.5, fmt.Errorf("failed to get source authority: %w", err)
	}

	// Weighted average of authority components
	return (authority.AuthorityScore*0.4 + authority.CredibilityScore*0.3 + authority.ReachScore*0.3), nil
}

// CalculateContentScore analyzes content for trending potential
func (s *EngagementService) CalculateContentScore(title, content, category string) float64 {
	score := 0.5 // Base score

	// Title analysis
	titleWords := strings.Fields(strings.ToLower(title))

	// Trending keywords (you can expand this list)
	trendingKeywords := map[string]float64{
		"breaking":   0.3,
		"urgent":     0.25,
		"exclusive":  0.2,
		"president":  0.15,
		"government": 0.1,
		"economy":    0.1,
		"crisis":     0.2,
		"election":   0.15,
		"scandal":    0.2,
		"victory":    0.1,
		"record":     0.1,
		"first":      0.1,
		"new":        0.05,
	}

	for _, word := range titleWords {
		if boost, exists := trendingKeywords[word]; exists {
			score += boost
		}
	}

	// Category boost
	categoryBoosts := map[string]float64{
		"politics":      0.2,
		"business":      0.15,
		"sports":        0.1,
		"technology":    0.1,
		"entertainment": 0.05,
	}

	if boost, exists := categoryBoosts[category]; exists {
		score += boost
	}

	// Content length factor (optimal length gets boost)
	contentLength := len(content)
	if contentLength > 500 && contentLength < 2000 {
		score += 0.1 // Sweet spot for engagement
	}

	return math.Min(score, 1.0) // Cap at 1.0
}

// CalculateAdvancedTrendingScore combines all factors
func (s *EngagementService) CalculateAdvancedTrendingScore(article models.Article, engagement models.ArticleEngagement, timeWindow string) (float64, error) {
	// 1. Engagement-Based Scoring (40% weight)
	totalEngagement := float64(engagement.Views*1 + engagement.Shares*3 + engagement.Comments*2 + engagement.Likes*2)
	engagementScore := math.Min(totalEngagement/100.0, 1.0) // Normalize to 0-1

	// 2. Velocity-Based Trending (30% weight)
	velocity, err := s.CalculateEngagementVelocity(article.ID, timeWindow)
	if err != nil {
		velocity = 0.0
	}
	velocityScore := math.Min(velocity/2.0, 1.0) // Normalize to 0-1

	// 3. Source Authority Weighting (20% weight)
	authorityScore, err := s.GetSourceAuthority(article.Source)
	if err != nil {
		authorityScore = 0.5
	}

	// 4. Time Decay Function (5% weight)
	recencyScore := s.CalculateRecencyScore(article.PublishedAt)

	// 5. Content Analysis (5% weight)
	contentScore := s.CalculateContentScore(article.Title, article.Content, article.Category)

	// Weighted combination
	finalScore := (engagementScore * 0.4) +
		(velocityScore * 0.3) +
		(authorityScore * 0.2) +
		(recencyScore * 0.05) +
		(contentScore * 0.05)

	return finalScore, nil
}

// UpdateSourceAuthority recalculates source authority based on performance
func (s *EngagementService) UpdateSourceAuthority(sourceName string) error {
	// Calculate metrics for the source
	var totalArticles int
	var avgEngagement float64

	err := s.db.QueryRow(`
		SELECT COUNT(a.id), 
			COALESCE(AVG(COALESCE(e.views, 0) + COALESCE(e.shares, 0) * 3 + 
						 COALESCE(e.comments, 0) * 2 + COALESCE(e.likes, 0) * 2), 0)
		FROM articles a
		LEFT JOIN article_engagement e ON a.id = e.article_id
		WHERE a.source = ?
			AND a.published_at >= DATETIME('now', '-30 days')
	`, sourceName).Scan(&totalArticles, &avgEngagement)

	if err != nil {
		return fmt.Errorf("failed to calculate source metrics: %w", err)
	}

	// Calculate authority components
	authorityScore := math.Min(avgEngagement/50.0, 1.0)             // Based on avg engagement
	credibilityScore := math.Min(float64(totalArticles)/100.0, 1.0) // Based on volume
	reachScore := (authorityScore + credibilityScore) / 2.0         // Combined metric

	// Update or insert source authority
	_, err = s.db.Exec(`
		INSERT INTO source_authority (source_name, authority_score, credibility_score, reach_score, total_articles, avg_engagement)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_name) DO UPDATE SET
			authority_score = ?,
			credibility_score = ?,
			reach_score = ?,
			total_articles = ?,
			avg_engagement = ?,
			last_calculated = CURRENT_TIMESTAMP
	`, sourceName, authorityScore, credibilityScore, reachScore, totalArticles, avgEngagement,
		authorityScore, credibilityScore, reachScore, totalArticles, avgEngagement)

	return err
}

// Helper functions
func (s *EngagementService) parseTimeWindow(timeWindow string) int {
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

func (s *EngagementService) getEventWeight(eventType string) float64 {
	weights := map[string]float64{
		"view":    1.0,
		"share":   3.0,
		"comment": 2.0,
		"like":    2.0,
	}

	if weight, exists := weights[eventType]; exists {
		return weight
	}
	return 1.0
}

// CacheTrendingScores pre-calculates and caches trending scores for performance
func (s *EngagementService) CacheTrendingScores(timeWindow string) error {
	// Clear old cache entries for this time window
	_, err := s.db.Exec(`
		DELETE FROM trending_cache 
		WHERE time_window = ? AND calculated_at < DATETIME('now', '-1 hour')
	`, timeWindow)
	if err != nil {
		return fmt.Errorf("failed to clear old cache: %w", err)
	}

	// Get articles to calculate trending scores for
	rows, err := s.db.Query(`
		SELECT a.id, a.title, a.content, a.url, a.author, a.source, 
			a.published_at, a.created_at, a.category, COALESCE(a.image_url, ''),
			COALESCE(e.views, 0), COALESCE(e.shares, 0), 
			COALESCE(e.comments, 0), COALESCE(e.likes, 0), 
			COALESCE(e.last_updated, a.created_at)
		FROM articles a
		LEFT JOIN article_engagement e ON a.id = e.article_id
		WHERE a.published_at >= DATETIME('now', '-' || ? || ' hours')
		ORDER BY a.published_at DESC
		LIMIT 1000
	`, s.parseTimeWindow(timeWindow))
	if err != nil {
		return fmt.Errorf("failed to query articles for trending cache: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var article models.Article
		var engagement models.ArticleEngagement

		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.URL,
			&article.Author, &article.Source, &article.PublishedAt, &article.CreatedAt,
			&article.Category, &article.ImageURL, &engagement.Views, &engagement.Shares,
			&engagement.Comments, &engagement.Likes, &engagement.LastUpdated)
		if err != nil {
			continue
		}

		engagement.ArticleID = article.ID

		// Calculate comprehensive trending score
		trendingScore, err := s.CalculateAdvancedTrendingScore(article, engagement, timeWindow)
		if err != nil {
			continue
		}

		// Calculate individual components for detailed analysis
		velocity, _ := s.CalculateEngagementVelocity(article.ID, timeWindow)
		recencyScore := s.CalculateRecencyScore(article.PublishedAt)
		authorityScore, _ := s.GetSourceAuthority(article.Source)
		contentScore := s.CalculateContentScore(article.Title, article.Content, article.Category)

		// Cache the results
		_, err = s.db.Exec(`
			INSERT OR REPLACE INTO trending_cache 
			(article_id, time_window, trending_score, engagement_velocity, 
			 recency_score, authority_score, content_score)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, article.ID, timeWindow, trendingScore, velocity, recencyScore, authorityScore, contentScore)
		if err != nil {
			continue // Log error in production
		}
	}

	return nil
}
