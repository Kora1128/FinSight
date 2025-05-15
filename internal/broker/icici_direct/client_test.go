package icici_direct

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	apiSecret := "test-api-secret"
	client := NewClient(apiKey, apiSecret)
	assert.NotNil(t, client)
	assert.Equal(t, apiKey, client.apiKey)
	assert.Equal(t, apiSecret, client.apiSecret)
}

func TestLogin(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	// Test with valid credentials
	err := client.Login("valid-token", "valid-secret")
	assert.NoError(t, err)

	// Test with invalid credentials
	err = client.Login("", "")
	assert.Error(t, err)
}

func TestGetHoldings(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	holdings, err := client.GetHoldings(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, holdings)
	assert.Empty(t, holdings) // Currently returns an empty slice
}

func TestGetPositions(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	positions, err := client.GetPositions(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, positions)
	assert.Empty(t, positions) // Currently returns an empty slice
}
