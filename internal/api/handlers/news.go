package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Kora1128/FinSight/internal/news"
	"github.com/gin-gonic/gin"
)

// Common errors
var (
	ErrInvalidLimit   = errors.New("invalid limit parameter")
	ErrMissingSymbol  = errors.New("stock symbol is required")
	ErrInvalidSource  = errors.New("invalid source configuration")
	ErrSourceExists   = errors.New("source already exists")
	ErrSourceNotFound = errors.New("source not found")
	ErrInvalidRequest = errors.New("invalid request")
)

// NewsHandler handles news-related HTTP requests
type NewsHandler struct {
	processor *news.Processor
	fetcher   *news.NewsFetcher
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(processor *news.Processor, fetcher *news.NewsFetcher) *NewsHandler {
	return &NewsHandler{
		processor: processor,
		fetcher:   fetcher,
	}
}

// GetRecommendations returns all recommendations
func (h *NewsHandler) GetRecommendations(c *gin.Context) {
	recommendations := h.processor.GetLatestRecommendations(100) // Limit to 100 recommendations
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   recommendations,
	})
}

// GetLatestRecommendations returns the most recent recommendations
func (h *NewsHandler) GetLatestRecommendations(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  ErrInvalidLimit.Error(),
		})
		return
	}

	if limit <= 0 || limit > 100 {
		limit = 20 // Default to 10 if limit is invalid
	}

	recommendations := h.processor.GetLatestRecommendations(limit)
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   recommendations,
	})
}

// GetRecommendationsByStock returns recommendations for a specific stock
func (h *NewsHandler) GetRecommendationsByStock(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  ErrMissingSymbol.Error(),
		})
		return
	}

	recommendations := h.processor.GetRecommendationsByStock(symbol)
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   recommendations,
	})
}

// GetSources returns all configured news sources
func (h *NewsHandler) GetSources(c *gin.Context) {
	sources := h.fetcher.GetSources()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sources,
	})
}

// AddSourceRequest represents the request body for adding a new source
type AddSourceRequest struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required,url"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category" binding:"required"`
}

// AddSource adds a new news source
func (h *NewsHandler) AddSource(c *gin.Context) {
	var req AddSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  ErrInvalidRequest.Error(),
		})
		return
	}

	source := news.Source{
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		Category:    req.Category,
	}

	if err := h.fetcher.AddSource(source); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, news.ErrSourceExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   source,
	})
}

// RemoveSource removes a news source
func (h *NewsHandler) RemoveSource(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  ErrMissingSymbol.Error(),
		})
		return
	}

	if err := h.fetcher.RemoveSource(name); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, news.ErrSourceNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Source removed successfully",
	})
}
