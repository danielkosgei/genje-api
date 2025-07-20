package services

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

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

	// Use advanced NLP-based summarization
	summary := s.generateNLPSummary(sentences, title)

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

// generateNLPSummary creates an advanced NLP-based summary using TF-IDF and multiple ranking algorithms
func (s *SummarizerService) generateNLPSummary(sentences []string, title string) string {
	if len(sentences) == 0 {
		return ""
	}

	// Calculate TF-IDF scores for all sentences
	tfidfScores := s.calculateTFIDF(sentences)

	// Calculate multiple scoring metrics
	scoredSentences := s.calculateAdvancedScores(sentences, title, tfidfScores)

	// Apply sentence clustering to avoid redundancy
	clusteredSentences := s.clusterSentences(scoredSentences)

	// Select best sentences using multiple criteria
	summary := s.selectOptimalSentences(clusteredSentences, 300) // Target ~300 chars

	return summary
}

// calculateTFIDF computes TF-IDF scores for all terms in all sentences
func (s *SummarizerService) calculateTFIDF(sentences []string) map[string]map[string]float64 {
	// Build vocabulary and document frequency
	vocabulary := make(map[string]int)
	termFreq := make(map[string]map[string]int)

	for i, sentence := range sentences {
		words := s.tokenize(strings.ToLower(sentence))
		termFreq[fmt.Sprintf("sent_%d", i)] = make(map[string]int)

		for _, word := range words {
			if s.isValidTerm(word) {
				vocabulary[word]++
				termFreq[fmt.Sprintf("sent_%d", i)][word]++
			}
		}
	}

	// Calculate TF-IDF scores
	tfidfScores := make(map[string]map[string]float64)
	numDocs := float64(len(sentences))

	for i, sentence := range sentences {
		sentKey := fmt.Sprintf("sent_%d", i)
		tfidfScores[sentKey] = make(map[string]float64)
		words := s.tokenize(strings.ToLower(sentence))
		totalWords := len(words)

		for _, word := range words {
			if s.isValidTerm(word) {
				// Term Frequency
				tf := float64(termFreq[sentKey][word]) / float64(totalWords)

				// Inverse Document Frequency
				idf := math.Log(numDocs / float64(vocabulary[word]))

				// TF-IDF Score
				tfidfScores[sentKey][word] = tf * idf
			}
		}
	}

	return tfidfScores
}

// calculateAdvancedScores uses multiple NLP techniques to score sentences
func (s *SummarizerService) calculateAdvancedScores(sentences []string, title string, tfidfScores map[string]map[string]float64) []AdvancedScoredSentence {
	var scored []AdvancedScoredSentence
	titleWords := s.tokenize(strings.ToLower(title))

	for i, sentence := range sentences {
		sentKey := fmt.Sprintf("sent_%d", i)
		words := s.tokenize(strings.ToLower(sentence))

		// 1. TF-IDF Score (semantic importance)
		tfidfScore := 0.0
		for _, word := range words {
			if score, exists := tfidfScores[sentKey][word]; exists {
				tfidfScore += score
			}
		}
		tfidfScore = tfidfScore / float64(len(words)) // Normalize by sentence length

		// 2. Position Score (journalism principle: important info first)
		positionScore := math.Exp(-float64(i) / 3.0) // Exponential decay

		// 3. Title Similarity Score (cosine similarity)
		titleSimilarity := s.calculateCosineSimilarity(words, titleWords)

		// 4. Sentence Length Score (prefer medium-length sentences)
		lengthScore := s.calculateLengthScore(len(words))

		// 5. Named Entity Score (sentences with names, places, organizations)
		entityScore := s.calculateEntityScore(sentence)

		// 6. Numerical Information Score (dates, numbers, statistics)
		numericalScore := s.calculateNumericalScore(sentence)

		// 7. Centrality Score (similarity to other sentences)
		centralityScore := s.calculateCentralityScore(i, sentences)

		// 8. Discourse Markers Score (sentences with "however", "therefore", etc.)
		discourseScore := s.calculateDiscourseScore(sentence)

		// Weighted combination of all scores
		finalScore := (tfidfScore * 0.25) +
			(positionScore * 0.20) +
			(titleSimilarity * 0.15) +
			(lengthScore * 0.10) +
			(entityScore * 0.10) +
			(numericalScore * 0.08) +
			(centralityScore * 0.07) +
			(discourseScore * 0.05)

		scored = append(scored, AdvancedScoredSentence{
			Text:            sentence,
			Score:           finalScore,
			Index:           i,
			TFIDFScore:      tfidfScore,
			PositionScore:   positionScore,
			TitleSimilarity: titleSimilarity,
			LengthScore:     lengthScore,
			EntityScore:     entityScore,
			NumericalScore:  numericalScore,
			CentralityScore: centralityScore,
			DiscourseScore:  discourseScore,
		})
	}

	// Sort by final score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	return scored
}

// tokenize splits text into meaningful tokens
func (s *SummarizerService) tokenize(text string) []string {
	// Remove punctuation and split
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	cleaned := reg.ReplaceAllString(text, " ")
	words := strings.Fields(cleaned)

	var tokens []string
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 && s.isValidTerm(word) {
			tokens = append(tokens, word)
		}
	}
	return tokens
}

// isValidTerm checks if a term should be included in analysis
func (s *SummarizerService) isValidTerm(word string) bool {
	// Enhanced stop words list including Swahili
	stopWords := map[string]bool{
		// English stop words
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "them": true,
		"their": true, "there": true, "where": true, "when": true, "why": true, "how": true,
		"can": true, "may": true, "might": true, "must": true, "shall": true, "about": true,
		"into": true, "through": true, "during": true, "before": true, "after": true,
		"above": true, "below": true, "up": true, "down": true, "out": true, "off": true,
		"over": true, "under": true, "again": true, "further": true, "then": true, "once": true,

		// Swahili stop words
		"na": true, "ya": true, "wa": true, "ni": true, "za": true, "la": true, "kwa": true,
		"katika": true, "hii": true, "hiyo": true, "hizo": true, "haya": true, "hao": true,
		"yeye": true, "mimi": true, "wewe": true, "sisi": true, "ninyi": true, "wao": true,
		"wake": true, "wako": true, "wangu": true, "wetu": true, "wenu": true, "zao": true,
		"kuwa": true, "kama": true, "lakini": true, "au": true, "ama": true, "bali": true,
		"pia": true, "tu": true, "kwamba": true, "pale": true, "hapa": true, "hapo": true,
	}

	return len(word) >= 3 && !stopWords[word] && unicode.IsLetter(rune(word[0]))
}

// calculateCosineSimilarity computes cosine similarity between two word lists
func (s *SummarizerService) calculateCosineSimilarity(words1, words2 []string) float64 {
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Create word frequency maps
	freq1 := make(map[string]int)
	freq2 := make(map[string]int)

	for _, word := range words1 {
		freq1[word]++
	}
	for _, word := range words2 {
		freq2[word]++
	}

	// Calculate dot product and magnitudes
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	allWords := make(map[string]bool)
	for word := range freq1 {
		allWords[word] = true
	}
	for word := range freq2 {
		allWords[word] = true
	}

	for word := range allWords {
		f1 := float64(freq1[word])
		f2 := float64(freq2[word])

		dotProduct += f1 * f2
		magnitude1 += f1 * f1
		magnitude2 += f2 * f2
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

// calculateLengthScore gives optimal score for medium-length sentences
func (s *SummarizerService) calculateLengthScore(wordCount int) float64 {
	// Optimal sentence length is around 15-25 words
	if wordCount < 8 {
		return 0.3 // Too short
	} else if wordCount > 40 {
		return 0.4 // Too long
	} else if wordCount >= 12 && wordCount <= 28 {
		return 1.0 // Optimal range
	} else {
		return 0.7 // Acceptable
	}
}

// calculateEntityScore identifies sentences with named entities
func (s *SummarizerService) calculateEntityScore(sentence string) float64 {
	score := 0.0

	// Look for capitalized words (potential named entities)
	words := strings.Fields(sentence)
	capitalizedCount := 0

	for _, word := range words {
		// Remove punctuation
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}…")
		if len(cleanWord) > 2 && unicode.IsUpper(rune(cleanWord[0])) {
			capitalizedCount++
		}
	}

	// Score based on density of capitalized words
	if len(words) > 0 {
		density := float64(capitalizedCount) / float64(len(words))
		score = math.Min(density*2.0, 1.0) // Cap at 1.0
	}

	// Bonus for specific entity patterns
	entityPatterns := []string{
		"President", "Minister", "CEO", "Director", "Chairman",
		"Kenya", "Nairobi", "Mombasa", "Kisumu", "Nakuru",
		"KSh", "USD", "million", "billion", "percent", "%",
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	sentenceLower := strings.ToLower(sentence)
	for _, pattern := range entityPatterns {
		if strings.Contains(sentenceLower, strings.ToLower(pattern)) {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// calculateNumericalScore identifies sentences with numerical information
func (s *SummarizerService) calculateNumericalScore(sentence string) float64 {
	score := 0.0

	// Look for numbers, dates, percentages, currency
	numericalPatterns := []string{
		`\d+`, `\d+\.\d+`, `\d+%`, `\d+,\d+`,
		`KSh\s*\d+`, `USD\s*\d+`, `\$\d+`,
		`20\d{2}`, `19\d{2}`, // Years
		`\d{1,2}/\d{1,2}/\d{4}`, `\d{1,2}-\d{1,2}-\d{4}`, // Dates
	}

	for _, pattern := range numericalPatterns {
		if matched, _ := regexp.MatchString(pattern, sentence); matched {
			score += 0.2
		}
	}

	return math.Min(score, 1.0)
}

// calculateCentralityScore measures how similar a sentence is to others
func (s *SummarizerService) calculateCentralityScore(index int, sentences []string) float64 {
	if len(sentences) <= 1 {
		return 0.0
	}

	currentWords := s.tokenize(strings.ToLower(sentences[index]))
	totalSimilarity := 0.0

	for i, otherSentence := range sentences {
		if i != index {
			otherWords := s.tokenize(strings.ToLower(otherSentence))
			similarity := s.calculateCosineSimilarity(currentWords, otherWords)
			totalSimilarity += similarity
		}
	}

	return totalSimilarity / float64(len(sentences)-1)
}

// calculateDiscourseScore identifies sentences with discourse markers
func (s *SummarizerService) calculateDiscourseScore(sentence string) float64 {
	score := 0.0
	sentenceLower := strings.ToLower(sentence)

	// Discourse markers that indicate important information
	discourseMarkers := []string{
		"however", "therefore", "furthermore", "moreover", "consequently",
		"nevertheless", "nonetheless", "meanwhile", "subsequently", "finally",
		"in conclusion", "as a result", "on the other hand", "in addition",
		"for example", "for instance", "specifically", "particularly",
		"according to", "reported", "announced", "revealed", "confirmed",
		"stated", "declared", "emphasized", "highlighted", "noted",
		// Swahili discourse markers
		"hata hivyo", "kwa hiyo", "zaidi ya hayo", "aidha", "hatimaye",
		"kwa mfano", "haswa", "kulingana na", "alisema", "alitangaza",
		"aliripoti", "alifichua", "alisisitiza", "aliweka wazi",
	}

	for _, marker := range discourseMarkers {
		if strings.Contains(sentenceLower, marker) {
			score += 0.3
		}
	}

	return math.Min(score, 1.0)
}

// clusterSentences groups similar sentences to avoid redundancy
func (s *SummarizerService) clusterSentences(sentences []AdvancedScoredSentence) []AdvancedScoredSentence {
	if len(sentences) <= 2 {
		return sentences
	}

	// Simple clustering: remove sentences that are too similar to higher-scored ones
	var filtered []AdvancedScoredSentence
	similarityThreshold := 0.7

	for i, sentence := range sentences {
		isUnique := true
		sentenceWords := s.tokenize(strings.ToLower(sentence.Text))

		// Check against already selected sentences
		for _, selected := range filtered {
			selectedWords := s.tokenize(strings.ToLower(selected.Text))
			similarity := s.calculateCosineSimilarity(sentenceWords, selectedWords)

			if similarity > similarityThreshold {
				isUnique = false
				break
			}
		}

		if isUnique || i < 2 { // Always keep top 2 sentences
			filtered = append(filtered, sentence)
		}
	}

	return filtered
}

// selectOptimalSentences chooses the best sentences for the final summary
func (s *SummarizerService) selectOptimalSentences(sentences []AdvancedScoredSentence, targetLength int) string {
	if len(sentences) == 0 {
		return ""
	}

	var selected []AdvancedScoredSentence
	totalLength := 0

	// Greedy selection based on score and length constraints
	for _, sentence := range sentences {
		sentenceLength := len(sentence.Text)

		// Check if adding this sentence would exceed target length
		if totalLength+sentenceLength <= targetLength || len(selected) == 0 {
			selected = append(selected, sentence)
			totalLength += sentenceLength + 1 // +1 for space

			// Stop if we have enough content
			if len(selected) >= 3 || totalLength >= int(float64(targetLength)*0.8) {
				break
			}
		}
	}

	// Sort selected sentences by original order for better flow
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Index < selected[j].Index
	})

	// Build final summary
	var summaryParts []string
	for _, sentence := range selected {
		summaryParts = append(summaryParts, strings.TrimSpace(sentence.Text))
	}

	summary := strings.Join(summaryParts, " ")

	// Ensure summary doesn't exceed target length
	if len(summary) > targetLength {
		summary = summary[:targetLength-3] + "..."
	}

	return summary
}

// AdvancedScoredSentence represents a sentence with detailed NLP scoring
type AdvancedScoredSentence struct {
	Text            string
	Score           float64
	Index           int
	TFIDFScore      float64
	PositionScore   float64
	TitleSimilarity float64
	LengthScore     float64
	EntityScore     float64
	NumericalScore  float64
	CentralityScore float64
	DiscourseScore  float64
}

// ScoredSentence represents a sentence with its importance score
type ScoredSentence struct {
	Text  string
	Score float64
	Index int
}
