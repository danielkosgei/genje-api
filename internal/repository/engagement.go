package repository

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"genje-api/internal/models"
)

type EngagementRepository struct {
	db *sql.DB
}

func NewEngagementRepository(db *sql.DB) *EngagementRepository {
	return &EngagementRepository{db: db}
}

// TrackEngagement records an engagement event and updates counters
func (r *EngagementRepository) TrackEngagement(articleID int, req models.EngagementRequest) error {
	tx, err := r.db.Begin()
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

// GetArticleEngagement retrieves engagement metrics for an article
func (r *EngagementRepository) GetArticleEngagement(articleID int) (*models.ArticleEngagement, error) {
	var engagement models.ArticleEngagement
	err := r.db.QueryRow(`
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

// GetEngagementVelocity calculates engagement velocity over time windows
func (r *EngagementRepository) GetEngagementVelocity(articleID int, timeWindow string) (float64, error) {
	hours := 24
	switch timeWindow {
	case "1h":
		hours = 1
	case "6h":
		hours = 6
	case "12h":
		hours = 12
	case "24h":
		hours = 24
	case "7d":
		hours = 24 * 7
	}

	// Get engagement events in the time window
	var currentCount, previousCount int

	// Current period
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM engagement_events
		WHERE article_id = ? AND timestamp >= DATETIME('now', '-' || ? || ' hours')
	`, articleID, hours).Scan(&currentCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get current engagement count: %w", err)
	}

	// Previous period (same duration, earlier)
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM engagement_events
		WHERE article_id = ? 
		AND timestamp >= DATETIME('now', '-' || ? || ' hours')
		AND timestamp < DATETIME('now', '-' || ? || ' hours')
	`, articleID, hours*2, hours).Scan(&previousCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get previous engagement count: %w", err)
	}

	// Calculate velocity (rate of change)
	if previousCount == 0 {
		if currentCount > 0 {
			return 1.0, nil // High velocity for new engagement
		}
		return 0.0, nil
	}

	velocity := float64(currentCount-previousCount) / float64(previousCount)
	return math.Max(-1.0, math.Min(1.0, velocity)), nil // Clamp between -1 and 1
}

// UpdateSourceAuthority calculates and updates source authority scores
func (r *EngagementRepository) UpdateSourceAuthority(sourceName string) error {
	tx, err := r.db.Begin()
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

	// Calculate metrics for the source
	var totalArticles int
	var avgViews, avgShares, avgComments, avgLikes float64

	err = tx.QueryRow(`
		SELECT 
			COUNT(a.id) as total_articles,
			COALESCE(AVG(COALESCE(ae.views, 0)), 0) as avg_views,
			COALESCE(AVG(COALESCE(ae.shares, 0)), 0) as avg_shares,
			COALESCE(AVG(COALESCE(ae.comments, 0)), 0) as avg_comments,
			COALESCE(AVG(COALESCE(ae.likes, 0)), 0) as avg_likes
		FROM articles a
		LEFT JOIN article_engagement ae ON a.id = ae.article_id
		WHERE a.source = ?
	`, sourceName).Scan(&totalArticles, &avgViews, &avgShares, &avgComments, &avgLikes)
	if err != nil {
		return fmt.Errorf("failed to calculate source metrics: %w", err)
	}

	// Calculate authority scores
	authorityScore := r.calculateAuthorityScore(totalArticles, avgViews, avgShares)
	credibilityScore := r.calculateCredibilityScore(avgComments, avgLikes, avgShares)
	reachScore := r.calculateReachScore(avgViews, totalArticles)
	avgEngagement := (avgViews + avgShares*5 + avgComments*3 + avgLikes*2) / 4

	// Update or insert source authority
	_, err = tx.Exec(`
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
	if err != nil {
		return fmt.Errorf("failed to update source authority: %w", err)
	}

	return tx.Commit()
}

// Helper functions for authority scoring
func (r *EngagementRepository) calculateAuthorityScore(totalArticles int, avgViews, avgShares float64) float64 {
	// Normalize based on article count and engagement
	articleScore := math.Min(float64(totalArticles)/100.0, 1.0)      // Max at 100 articles
	engagementScore := math.Min((avgViews+avgShares*10)/1000.0, 1.0) // Weighted engagement
	return (articleScore*0.3 + engagementScore*0.7)                  // Favor engagement over volume
}

func (r *EngagementRepository) calculateCredibilityScore(avgComments, avgLikes, avgShares float64) float64 {
	// Higher comments and likes relative to shares indicate credibility
	if avgShares == 0 {
		return 0.5 // Neutral if no shares
	}
	ratio := (avgComments + avgLikes) / avgShares
	return math.Min(ratio/10.0, 1.0) // Normalize to 0-1
}

func (r *EngagementRepository) calculateReachScore(avgViews float64, totalArticles int) float64 {
	// Reach based on views per article
	if totalArticles == 0 {
		return 0.0
	}
	viewsPerArticle := avgViews / float64(totalArticles)
	return math.Min(viewsPerArticle/1000.0, 1.0) // Normalize to 0-1
}

// GetSourceAuthority retrieves authority scores for a source
func (r *EngagementRepository) GetSourceAuthority(sourceName string) (*models.SourceAuthority, error) {
	var authority models.SourceAuthority
	err := r.db.QueryRow(`
		SELECT id, source_name, authority_score, credibility_score, reach_score, 
			total_articles, avg_engagement, last_calculated
		FROM source_authority
		WHERE source_name = ?
	`, sourceName).Scan(&authority.ID, &authority.SourceName, &authority.AuthorityScore,
		&authority.CredibilityScore, &authority.ReachScore, &authority.TotalArticles,
		&authority.AvgEngagement, &authority.LastCalculated)

	if err == sql.ErrNoRows {
		// Return default scores if not found
		return &models.SourceAuthority{
			SourceName:       sourceName,
			AuthorityScore:   0.5,
			CredibilityScore: 0.5,
			ReachScore:       0.5,
			TotalArticles:    0,
			AvgEngagement:    0.0,
			LastCalculated:   time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get source authority: %w", err)
	}

	return &authority, nil
}

// GetTopEngagedArticles returns most engaged articles in time window
func (r *EngagementRepository) GetTopEngagedArticles(limit int, timeWindow string) ([]models.ArticleEngagement, error) {
	hours := 24
	switch timeWindow {
	case "1h":
		hours = 1
	case "6h":
		hours = 6
	case "12h":
		hours = 12
	case "24h":
		hours = 24
	case "7d":
		hours = 24 * 7
	}

	query := `
		SELECT ae.id, ae.article_id, ae.views, ae.shares, ae.comments, ae.likes, ae.last_updated
		FROM article_engagement ae
		JOIN articles a ON ae.article_id = a.id
		WHERE a.published_at >= DATETIME('now', '-' || ? || ' hours')
		ORDER BY (ae.views + ae.shares*5 + ae.comments*3 + ae.likes*2) DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top engaged articles: %w", err)
	}
	defer rows.Close()

	var engagements []models.ArticleEngagement
	for rows.Next() {
		var engagement models.ArticleEngagement
		err := rows.Scan(&engagement.ID, &engagement.ArticleID, &engagement.Views,
			&engagement.Shares, &engagement.Comments, &engagement.Likes, &engagement.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan engagement: %w", err)
		}
		engagements = append(engagements, engagement)
	}

	return engagements, nil
}
