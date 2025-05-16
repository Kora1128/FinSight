package news

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// StockResolver defines the interface for resolving stock symbols from text
type StockResolver interface {
	// ResolveSymbol extracts the most relevant stock symbol from the given text
	ResolveSymbol(ctx context.Context, text string) (string, error)
}

// OpenAIStockResolver implements StockResolver using OpenAI's GPT-3.5 Turbo
type OpenAIStockResolver struct {
	client *openai.Client
	cache  *RecommendationCache
	apiKey string
}

// NewOpenAIStockResolver creates a new OpenAI stock resolver
func NewOpenAIStockResolver(apiKey string) *OpenAIStockResolver {
	return &OpenAIStockResolver{
		client: openai.NewClient(apiKey),
		cache: NewRecommendationCache(CacheConfig{
			TTL:             24 * time.Hour,
			MaxItems:        1000,
			CleanupInterval: 1 * time.Hour,
		}),
		apiKey: apiKey,
	}
}

// ResolveSymbol implements the StockResolver interface
func (r *OpenAIStockResolver) ResolveSymbol(ctx context.Context, text string) (string, error) {
	// First check cache
	if symbol, found := r.cache.Get(text); found {
		return symbol.StockSymbol, nil
	}

	// Prepare the prompt
	prompt := fmt.Sprintf(`Extract the most relevant stock symbol from the following text. If it's about the general market, return 'NIFTY'. If no specific stock is mentioned, return an empty string. Only return the symbol, nothing else.

Text: %s`, text)

	// Make the request to OpenAI
	resp, err := r.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a financial assistant that extracts stock symbols from text. Only respond with the stock symbol, nothing else.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1, // Low temperature for more deterministic responses
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get completion from OpenAI: %w", err)
	}

	// Get the response
	symbol := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Cache the result if we got a valid symbol
	if symbol != "" {
		r.cache.Set(text, Recommendation{
			StockSymbol: symbol,
			CreatedAt:   time.Now(),
		})
	}

	return symbol, nil
}

// MockStockResolver implements StockResolver for testing
type MockStockResolver struct {
	Symbols map[string]string
}

// NewMockStockResolver creates a new mock stock resolver
func NewMockStockResolver() *MockStockResolver {
	return &MockStockResolver{
		Symbols: make(map[string]string),
	}
}

// ResolveSymbol implements the StockResolver interface for testing
func (r *MockStockResolver) ResolveSymbol(ctx context.Context, text string) (string, error) {
	// Check for exact matches first
	for symbol, pattern := range r.Symbols {
		if strings.Contains(strings.ToUpper(text), strings.ToUpper(pattern)) {
			return symbol, nil
		}
	}

	// Check for market-wide news
	marketKeywords := []string{"market", "sensex", "nifty", "bse", "nse", "stock market", "share market"}
	textLower := strings.ToLower(text)
	for _, keyword := range marketKeywords {
		if strings.Contains(textLower, keyword) {
			return "NIFTY", nil
		}
	}

	return "", nil
}
