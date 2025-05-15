package icici_direct

import (
	"context"
	"errors"
	"testing"

	"github.com/Kora1128/FinSight/internal/models"
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
	tests := []struct {
		name        string
		mockClient  *MockClient
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful login",
			mockClient:  NewMockClient(),
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "failed login",
			mockClient:  NewMockClient().WithLoginError(errors.New("invalid credentials")),
			wantErr:     true,
			expectedErr: errors.New("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mockClient.Login("test-token", "test-secret")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetHoldings(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *MockClient
		wantErr     bool
		expectedErr error
		expectedLen int
	}{
		{
			name:        "successful holdings fetch",
			mockClient:  NewMockClient().WithMockHoldings(GetDefaultMockHoldings()),
			wantErr:     false,
			expectedErr: nil,
			expectedLen: 2,
		},
		{
			name:        "failed holdings fetch",
			mockClient:  NewMockClient().WithHoldingsError(errors.New("api error")),
			wantErr:     true,
			expectedErr: errors.New("api error"),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			holdings, err := tt.mockClient.GetHoldings(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, holdings, tt.expectedLen)
				for _, holding := range holdings {
					assert.Equal(t, models.PlatformICICIDirect, holding.Platform)
					assert.Equal(t, models.HoldingTypeStock, holding.Type)
				}
			}
		})
	}
}

func TestGetPositions(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *MockClient
		wantErr     bool
		expectedErr error
		expectedLen int
	}{
		{
			name:        "successful positions fetch",
			mockClient:  NewMockClient().WithMockPositions(GetDefaultMockPositions()),
			wantErr:     false,
			expectedErr: nil,
			expectedLen: 1,
		},
		{
			name:        "failed positions fetch",
			mockClient:  NewMockClient().WithPositionsError(errors.New("api error")),
			wantErr:     true,
			expectedErr: errors.New("api error"),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positions, err := tt.mockClient.GetPositions(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, positions, tt.expectedLen)
				for _, position := range positions {
					assert.Equal(t, models.PlatformICICIDirect, position.Platform)
					assert.Equal(t, models.HoldingTypeStock, position.Type)
				}
			}
		})
	}
}
