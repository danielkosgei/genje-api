package services

import (
	"fmt"
	"strings"

	"genje-api/internal/repository"
)

type SummarizerService struct {
	articleRepo *repository.ArticleRepository
}

func NewSummarizerService(articleRepo *repository.ArticleRepository) *SummarizerService {
	return &SummarizerService{
		articleRepo: articleRepo,
	}
}

func (s *SummarizerService) SummarizeArticle(articleID int) (string, error) {
	article, err := s.articleRepo.GetArticleByID(articleID)
	if err != nil {
		return "", fmt.Errorf("failed to get article: %w", err)
	}

	if article == nil {
		return "", fmt.Errorf("article not found")
	}

	// If already summarized, return existing summary
	if article.Summary != "" {
		return article.Summary, nil
	}

	// Generate summary
	summary := s.generateSimpleSummary(article.Content)

	// Update database
	if err := s.articleRepo.UpdateSummary(articleID, summary); err != nil {
		return "", fmt.Errorf("failed to save summary: %w", err)
	}

	return summary, nil
}

func (s *SummarizerService) generateSimpleSummary(content string) string {
	if content == "" {
		return ""
	}

	// Remove HTML tags (simple approach)
	content = strings.ReplaceAll(content, "<", " <")
	content = strings.ReplaceAll(content, ">", "> ")
	
	// Simple sentence splitting
	sentences := strings.Split(content, ". ")
	
	// Take first 3 sentences
	if len(sentences) > 3 {
		sentences = sentences[:3]
	}

	summary := strings.Join(sentences, ". ")
	if !strings.HasSuffix(summary, ".") {
		summary += "."
	}

	// Clean up whitespace
	summary = strings.TrimSpace(summary)
	
	// Limit to 300 characters
	if len(summary) > 300 {
		summary = summary[:297] + "..."
	}

	return summary
} 