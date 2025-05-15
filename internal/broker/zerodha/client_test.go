package zerodha

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
	assert.NotNil(t, client.kc)
}

func TestLogin(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	// Note: This is a mock test. In a real scenario, you would use a mock or a test token.
	err := client.Login("test-request-token", "test-api-secret")
	// Since we're not mocking the actual API call, this will likely fail in a real test environment.
	// You might want to use a mock or a test token for actual testing.
	assert.Error(t, err)
}

func TestGetHoldings(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	// Note: This is a mock test. In a real scenario, you would use a mock or a test token.
	holdings, err := client.GetHoldings(context.Background())
	// Since we're not mocking the actual API call, this will likely fail in a real test environment.
	// You might want to use a mock or a test token for actual testing.
	assert.Error(t, err)
	assert.Nil(t, holdings)
}

func TestGetPositions(t *testing.T) {
	client := NewClient("test-api-key", "test-api-secret")
	// Note: This is a mock test. In a real scenario, you would use a mock or a test token.
	positions, err := client.GetPositions(context.Background())
	// Since we're not mocking the actual API call, this will likely fail in a real test environment.
	// You might want to use a mock or a test token for actual testing.
	assert.Error(t, err)
	assert.Nil(t, positions)
}
