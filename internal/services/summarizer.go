package services

import (
	"fmt"
	"regexp"
	"sort"
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
	if article.Summary != "" && strings.TrimSpace(article.Summary) != "" {
		return article.Summary, nil
	}

	// Generate summary
	summary := s.generateIntelligentSummary(article.Title, article.Content)

	// Update database
	if err := s.articleRepo.UpdateSummary(articleID, summary); err != nil {
		return "", fmt.Errorf("failed to save summary: %w", err)
	}

	return summary, nil
}

func (s *SummarizerService) generateIntelligentSummary(title, content string) string {
	if content == "" {
		return ""
	}

	// Clean and extract text from HTML
	cleanText := s.cleanHTML(content)
	if cleanText == "" {
		return ""
	}

	// Extract sentences
	sentences := s.extractSentences(cleanText)
	if len(sentences) == 0 {
		return ""
	}

	// Score sentences based on importance
	scoredSentences := s.scoreSentences(sentences, title)

	// Select top sentences for summary
	summary := s.selectTopSentences(scoredSentences, 2, 250)

	return summary
}

// cleanHTML removes HTML tags and extracts clean text
func (s *SummarizerService) cleanHTML(html string) string {
	// Remove script and style elements completely
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = scriptRegex.ReplaceAllString(html, "")
	html = styleRegex.ReplaceAllString(html, "")

	// Remove HTML tags but keep the content
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#8217;", "'")
	text = strings.ReplaceAll(text, "&#8220;", "\"")
	text = strings.ReplaceAll(text, "&#8221;", "\"")

	// Clean up whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// extractSentences splits text into meaningful sentences
func (s *SummarizerService) extractSentences(text string) []string {
	// Split on sentence endings
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	rawSentences := sentenceRegex.Split(text, -1)

	var sentences []string
	for _, sentence := range rawSentences {
		sentence = strings.TrimSpace(sentence)
		
		// Filter out very short or very long sentences
		if len(sentence) >= 20 && len(sentence) <= 500 {
			// Ensure sentence ends with punctuation
			if !strings.HasSuffix(sentence, ".") && !strings.HasSuffix(sentence, "!") && !strings.HasSuffix(sentence, "?") {
				sentence += "."
			}
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// scoreSentences assigns importance scores to sentences
func (s *SummarizerService) scoreSentences(sentences []string, title string) []ScoredSentence {
	titleWords := s.extractKeywords(strings.ToLower(title))
	
	var scored []ScoredSentence
	
	for i, sentence := range sentences {
		score := 0.0
		sentenceLower := strings.ToLower(sentence)
		words := strings.Fields(sentenceLower)

		// Position score - earlier sentences are more important
		positionScore := 1.0 - (float64(i) / float64(len(sentences)))
		score += positionScore * 0.3

		// Length score - prefer medium-length sentences
		lengthScore := 1.0
		if len(words) < 10 {
			lengthScore = 0.5
		} else if len(words) > 30 {
			lengthScore = 0.7
		}
		score += lengthScore * 0.2

		// Title relevance score
		titleMatches := 0
		for _, titleWord := range titleWords {
			if strings.Contains(sentenceLower, titleWord) {
				titleMatches++
			}
		}
		if len(titleWords) > 0 {
			titleRelevance := float64(titleMatches) / float64(len(titleWords))
			score += titleRelevance * 0.4
		}

		// Keyword density score
		keywordScore := s.calculateKeywordScore(sentenceLower)
		score += keywordScore * 0.1

		scored = append(scored, ScoredSentence{
			Text:  sentence,
			Score: score,
			Index: i,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	return scored
}

// extractKeywords extracts important words from text
func (s *SummarizerService) extractKeywords(text string) []string {
	// Common stop words to ignore
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "them": true,
		"their": true, "there": true, "where": true, "when": true, "why": true, "how": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}...")
		
		// Keep words that are not stop words and are meaningful length
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// calculateKeywordScore calculates importance based on keyword presence
func (s *SummarizerService) calculateKeywordScore(sentence string) float64 {
	// Look for important keywords that indicate key information
	importantKeywords := []string{
		"said", "announced", "reported", "according", "revealed", "confirmed",
		"government", "president", "minister", "official", "authority",
		"million", "billion", "percent", "increase", "decrease", "growth",
		"new", "first", "major", "significant", "important", "key",
	}

	score := 0.0
	for _, keyword := range importantKeywords {
		if strings.Contains(sentence, keyword) {
			score += 0.1
		}
	}

	return score
}

// selectTopSentences creates final summary from top-scored sentences
func (s *SummarizerService) selectTopSentences(scored []ScoredSentence, maxSentences int, maxLength int) string {
	if len(scored) == 0 {
		return ""
	}

	// Take top sentences but maintain original order
	selected := scored[:min(maxSentences, len(scored))]
	
	// Sort by original index to maintain reading flow
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Index < selected[j].Index
	})

	// Build summary
	var summaryParts []string
	totalLength := 0

	for _, sentence := range selected {
		if totalLength+len(sentence.Text) <= maxLength {
			summaryParts = append(summaryParts, sentence.Text)
			totalLength += len(sentence.Text) + 1 // +1 for space
		}
	}

	summary := strings.Join(summaryParts, " ")
	
	// Ensure we have at least one sentence
	if summary == "" && len(scored) > 0 {
		summary = scored[0].Text
		if len(summary) > maxLength {
			summary = summary[:maxLength-3] + "..."
		}
	}

	return summary
}

// ScoredSentence represents a sentence with its importance score
type ScoredSentence struct {
	Text  string
	Score float64
	Index int
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 