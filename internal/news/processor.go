package news

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Processor handles filtering and processing of news items
type Processor struct {
	cache         *RecommendationCache
	stockResolver StockResolver
}

// NewProcessor creates a new news processor
func NewProcessor(cache *RecommendationCache, openAIKey string) *Processor {
	return &Processor{
		cache:         cache,
		stockResolver: NewOpenAIStockResolver(openAIKey),
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
		if recommendation.Confidence > 0.5 { // Only keep high confidence recommendations
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

	// Count positive and negative matches
	positiveCount := 0
	negativeCount := 0

	// Check title and description for keywords
	text := title + " " + description
	for _, keyword := range PositiveKeywords {
		if strings.Contains(text, keyword) {
			positiveCount++
		}
	}
	for _, keyword := range NegativeKeywords {
		if strings.Contains(text, keyword) {
			negativeCount++
		}
	}

	// Calculate sentiment score (-1 to 1)
	totalCount := positiveCount + negativeCount
	if totalCount == 0 {
		return NeutralSentimentScore // Neutral if no keywords found
	}

	// Normalize to -1 to 1 range
	sentiment := float64(positiveCount-negativeCount) / float64(totalCount)

	// Adjust based on source reliability
	switch item.Source {
	case "MoneyControl":
		sentiment *= MoneyControlMultiplier
	case "Economic Times":
		sentiment *= EconomicTimesMultiplier
	case "Business Standard":
		sentiment *= BusinessStandardMultiplier
	}

	// Ensure sentiment stays within bounds
	if sentiment > MaxSentimentScore {
		sentiment = MaxSentimentScore
	} else if sentiment < MinSentimentScore {
		sentiment = MinSentimentScore
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

	// Extract stock symbol from title or description
	stockSymbol := p.extractStockSymbol(item)

	return Recommendation{
		StockSymbol: stockSymbol,
		Action:      action,
		Confidence:  confidence,
		Reason:      p.generateReason(item, action, confidence),
		NewsItem:    item,
		CreatedAt:   time.Now(),
	}
}

// calculateRelevanceScore calculates how relevant a news item is for investment decisions
func (p *Processor) calculateRelevanceScore(item NewsItem) float64 {
	var score float64

	// Check for important keywords
	title := strings.ToLower(item.Title)
	description := strings.ToLower(item.Description)

	for _, keyword := range RelevanceKeywords {
		if strings.Contains(title, keyword) || strings.Contains(description, keyword) {
			score += KeywordMatchScore
		}
	}

	// Check source reliability
	if ReliableSources[item.Source] {
		score += SourceReliabilityScore
	}

	// Check recency
	age := time.Since(item.PublishedAt)
	if age < time.Duration(RecentNewsThreshold)*time.Second {
		score += RecentNewsScore
	} else if age < time.Duration(OlderNewsThreshold)*time.Second {
		score += OlderNewsScore
	}

	// Normalize score to 0-1 range
	if score > MaxRelevanceScore {
		score = MaxRelevanceScore
	}

	return score
}

// determineAction determines the recommended action based on sentiment and relevance
func (p *Processor) determineAction(sentiment, relevance float64) string {
	if relevance < RelevanceThreshold {
		return ActionWatch
	}

	if sentiment > PositiveSentimentThreshold {
		return ActionBuy
	} else if sentiment < NegativeSentimentThreshold {
		return ActionSell
	} else {
		return ActionHold
	}
}

// calculateConfidence calculates the confidence level in the recommendation
func (p *Processor) calculateConfidence(item NewsItem) float64 {
	var confidence float64

	// Source reliability
	switch item.Source {
	case "MoneyControl":
		confidence += MoneyControlConfidence
	case "Economic Times":
		confidence += EconomicTimesConfidence
	case "Business Standard":
		confidence += BusinessStandardConfidence
	default:
		confidence += DefaultSourceConfidence
	}

	// Content quality
	if len(item.Description) > 100 {
		confidence += ContentQualityScore
	}

	// Sentiment strength
	sentimentStrength := item.Sentiment
	if sentimentStrength < 0 {
		sentimentStrength = -sentimentStrength
	}
	confidence += sentimentStrength * SentimentStrengthWeight

	// Normalize confidence to 0-1 range
	if confidence > MaxConfidenceScore {
		confidence = MaxConfidenceScore
	}

	return confidence
}

// extractStockSymbol extracts the stock symbol from the news item
func (p *Processor) extractStockSymbol(item NewsItem) string {
	// Combine title and description for better context
	text := item.Title + " " + item.Description

	// Use the stock resolver to get the symbol
	symbol, err := p.stockResolver.ResolveSymbol(context.Background(), text)
	if err != nil {
		// Log the error but continue processing
		// In a production environment, you might want to use proper logging
		fmt.Printf("Error resolving stock symbol: %v\n", err)
		return ""
	}

	return symbol
}

// generateReason generates a human-readable reason for the recommendation
func (p *Processor) generateReason(item NewsItem, action string, confidence float64) string {
	var reason strings.Builder

	reason.WriteString("Based on ")
	if confidence > 0.8 {
		reason.WriteString("strong ")
	} else if confidence > 0.5 {
		reason.WriteString("moderate ")
	} else {
		reason.WriteString("weak ")
	}

	reason.WriteString("sentiment from ")
	reason.WriteString(item.Source)
	reason.WriteString(" news: ")
	reason.WriteString(item.Title)

	return reason.String()
}

// GetRecommendationsByStock returns recommendations for a specific stock
func (p *Processor) GetRecommendationsByStock(stockSymbol string) []Recommendation {
	allRecs := p.cache.GetAll()
	var stockRecs []Recommendation

	for _, rec := range allRecs {
		if rec.StockSymbol == stockSymbol {
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

	// Sort by CreatedAt
	for i := 0; i < len(allRecs)-1; i++ {
		for j := i + 1; j < len(allRecs); j++ {
			if allRecs[i].CreatedAt.Before(allRecs[j].CreatedAt) {
				allRecs[i], allRecs[j] = allRecs[j], allRecs[i]
			}
		}
	}

	return allRecs[:limit]
}
