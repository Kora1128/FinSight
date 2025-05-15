package news

import (
	"context"
	"strings"
	"time"
)

// Processor handles filtering and processing of news items
type Processor struct {
	cache *RecommendationCache
}

// NewProcessor creates a new news processor
func NewProcessor(cache *RecommendationCache) *Processor {
	return &Processor{
		cache: cache,
	}
}

// ProcessNews processes a list of news items and returns recommendations
func (p *Processor) ProcessNews(ctx context.Context, newsItems []NewsItem) []Recommendation {
	var recommendations []Recommendation

	for _, item := range newsItems {
		// Skip if already in cache
		if _, found := p.cache.Get(item.Link); found {
			continue
		}

		// Process the news item
		recommendation := p.processNewsItem(item)
		if recommendation.RelevanceScore > 0.5 { // Only keep relevant items
			recommendations = append(recommendations, recommendation)
			p.cache.Set(item.Link, recommendation)
		}
	}

	return recommendations
}

// analyzeSentiment performs basic sentiment analysis on a news item
func (p *Processor) analyzeSentiment(item NewsItem) float64 {
	// Convert text to lowercase for case-insensitive matching
	title := strings.ToLower(item.Title)
	description := strings.ToLower(item.Description)

	// Define positive and negative keywords
	positiveKeywords := []string{
		"strong", "growth", "profit", "gain", "upgrade", "positive", "bullish",
		"increase", "higher", "better", "exceed", "beat", "surge", "rise",
		"outperform", "success", "opportunity", "potential", "promising",
	}

	negativeKeywords := []string{
		"weak", "loss", "decline", "downgrade", "negative", "bearish",
		"decrease", "lower", "worse", "miss", "fall", "drop", "underperform",
		"risk", "concern", "warning", "caution", "volatile",
	}

	// Count positive and negative matches
	positiveCount := 0
	negativeCount := 0

	// Check title and description for keywords
	text := title + " " + description
	for _, keyword := range positiveKeywords {
		if strings.Contains(text, keyword) {
			positiveCount++
		}
	}
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			negativeCount++
		}
	}

	// Calculate sentiment score (-1 to 1)
	totalCount := positiveCount + negativeCount
	if totalCount == 0 {
		return 0 // Neutral if no keywords found
	}

	// Normalize to -1 to 1 range
	sentiment := float64(positiveCount-negativeCount) / float64(totalCount)

	// Adjust based on source reliability
	switch item.Source.Name {
	case "MoneyControl":
		sentiment *= 1.2 // Boost sentiment for reliable sources
	case "Economic Times":
		sentiment *= 1.1
	case "Business Standard":
		sentiment *= 1.1
	}

	// Ensure sentiment stays within -1 to 1 range
	if sentiment > 1.0 {
		sentiment = 1.0
	} else if sentiment < -1.0 {
		sentiment = -1.0
	}

	return sentiment
}

// processNewsItem processes a single news item and returns a recommendation
func (p *Processor) processNewsItem(item NewsItem) Recommendation {
	// Calculate sentiment if not already set
	if item.Sentiment == 0 {
		item.Sentiment = p.analyzeSentiment(item)
	}

	// Calculate relevance score based on various factors
	relevanceScore := p.calculateRelevanceScore(item)

	// Determine action based on sentiment and relevance
	action := p.determineAction(item.Sentiment, relevanceScore)

	// Calculate confidence based on source reliability and content quality
	confidence := p.calculateConfidence(item)

	return Recommendation{
		NewsItem:       item,
		RelevanceScore: relevanceScore,
		Action:         action,
		Confidence:     confidence,
		LastUpdated:    time.Now(),
	}
}

// calculateRelevanceScore calculates how relevant a news item is for investment decisions
func (p *Processor) calculateRelevanceScore(item NewsItem) float64 {
	var score float64

	// Check for important keywords
	keywords := []string{
		"earnings", "quarterly results", "financial results",
		"dividend", "acquisition", "merger", "takeover",
		"upgrade", "downgrade", "analyst", "rating",
		"guidance", "forecast", "outlook", "bullish", "bearish",
	}

	title := strings.ToLower(item.Title)
	description := strings.ToLower(item.Description)

	for _, keyword := range keywords {
		if strings.Contains(title, keyword) || strings.Contains(description, keyword) {
			score += 0.2
		}
	}

	// Check source reliability
	switch item.Source.Name {
	case "MoneyControl", "Economic Times", "Business Standard":
		score += 0.3
	}

	// Check recency
	age := time.Since(item.PublishedAt)
	if age < 24*time.Hour {
		score += 0.2
	} else if age < 48*time.Hour {
		score += 0.1
	}

	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// determineAction determines the recommended action based on sentiment and relevance
func (p *Processor) determineAction(sentiment, relevance float64) string {
	if relevance < 0.5 {
		return "WATCH"
	}

	if sentiment > 0.3 {
		return "BUY"
	} else if sentiment < -0.3 {
		return "SELL"
	} else {
		return "HOLD"
	}
}

// calculateConfidence calculates the confidence level in the recommendation
func (p *Processor) calculateConfidence(item NewsItem) float64 {
	var confidence float64

	// Source reliability
	switch item.Source.Name {
	case "MoneyControl":
		confidence += 0.4
	case "Economic Times":
		confidence += 0.35
	case "Business Standard":
		confidence += 0.35
	default:
		confidence += 0.2
	}

	// Content quality
	if len(item.Description) > 100 {
		confidence += 0.2
	}

	// Sentiment strength
	sentimentStrength := item.Sentiment
	if sentimentStrength < 0 {
		sentimentStrength = -sentimentStrength
	}
	confidence += sentimentStrength * 0.2

	// Normalize confidence to 0-1 range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// GetRecommendationsByStock returns recommendations for a specific stock
func (p *Processor) GetRecommendationsByStock(stockSymbol string) []Recommendation {
	allRecs := p.cache.GetAll()
	var stockRecs []Recommendation

	for _, rec := range allRecs {
		if strings.Contains(strings.ToLower(rec.Title), strings.ToLower(stockSymbol)) ||
			strings.Contains(strings.ToLower(rec.Description), strings.ToLower(stockSymbol)) {
			stockRecs = append(stockRecs, rec)
		}
	}

	return stockRecs
}

// GetLatestRecommendations returns the most recent recommendations
func (p *Processor) GetLatestRecommendations(limit int) []Recommendation {
	allRecs := p.cache.GetAll()
	if len(allRecs) <= limit {
		return allRecs
	}

	// Sort by LastUpdated
	for i := 0; i < len(allRecs)-1; i++ {
		for j := i + 1; j < len(allRecs); j++ {
			if allRecs[i].LastUpdated.Before(allRecs[j].LastUpdated) {
				allRecs[i], allRecs[j] = allRecs[j], allRecs[i]
			}
		}
	}

	return allRecs[:limit]
}
