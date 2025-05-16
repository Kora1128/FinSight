package news

// SentimentAnalysis constants
const (
	// Sentiment score bounds
	MaxSentimentScore     = 1.0
	MinSentimentScore     = -1.0
	NeutralSentimentScore = 0.0

	// Sentiment thresholds for actions
	PositiveSentimentThreshold = 0.3
	NegativeSentimentThreshold = -0.3

	// Source reliability multipliers
	MoneyControlMultiplier     = 1.2
	EconomicTimesMultiplier    = 1.1
	BusinessStandardMultiplier = 1.1
)

// RelevanceScore constants
const (
	// Relevance score bounds
	MaxRelevanceScore  = 1.0
	MinRelevanceScore  = 0.0
	RelevanceThreshold = 0.5

	// Relevance score components
	KeywordMatchScore      = 0.2
	SourceReliabilityScore = 0.3
	RecentNewsScore        = 0.2
	OlderNewsScore         = 0.1

	// Time thresholds for recency
	RecentNewsThreshold = 24 * 60 * 60 // 24 hours in seconds
	OlderNewsThreshold  = 48 * 60 * 60 // 48 hours in seconds
)

// ConfidenceScore constants
const (
	// Confidence score bounds
	MaxConfidenceScore = 1.0
	MinConfidenceScore = 0.0

	// Confidence score components
	MoneyControlConfidence     = 0.4
	EconomicTimesConfidence    = 0.35
	BusinessStandardConfidence = 0.35
	DefaultSourceConfidence    = 0.2
	ContentQualityScore        = 0.2
	SentimentStrengthWeight    = 0.2
)

// Action types
const (
	ActionBuy   = "BUY"
	ActionSell  = "SELL"
	ActionHold  = "HOLD"
	ActionWatch = "WATCH"
)

// Keywords for sentiment analysis
var (
	// Positive keywords indicate bullish or positive sentiment
	PositiveKeywords = []string{
		"strong", "growth", "profit", "gain", "upgrade", "positive", "bullish",
		"increase", "higher", "better", "exceed", "beat", "surge", "rise",
		"outperform", "success", "opportunity", "potential", "promising",
		"dividend", "acquisition", "expansion", "record", "breakthrough",
		"innovation", "partnership", "award", "recognition", "milestone",
	}

	// Negative keywords indicate bearish or negative sentiment
	NegativeKeywords = []string{
		"weak", "loss", "decline", "downgrade", "negative", "bearish",
		"decrease", "lower", "worse", "miss", "fall", "drop", "underperform",
		"risk", "concern", "warning", "caution", "volatile", "uncertainty",
		"challenge", "pressure", "decline", "reduction", "cut", "delay",
		"disappoint", "struggle", "difficulty", "setback",
	}

	// Relevance keywords indicate important financial news
	RelevanceKeywords = []string{
		"earnings", "quarterly results", "financial results",
		"dividend", "acquisition", "merger", "takeover",
		"upgrade", "downgrade", "analyst", "rating",
		"guidance", "forecast", "outlook", "target price",
		"revenue", "profit", "margin", "growth",
		"strategy", "plan", "initiative", "investment",
		"partnership", "agreement", "contract", "deal",
	}
)

// Source reliability rankings
var (
	// ReliableSources contains names of sources considered highly reliable
	ReliableSources = map[string]bool{
		"MoneyControl":      true,
		"Economic Times":    true,
		"Business Standard": true,
	}
)
