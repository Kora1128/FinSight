package broker

import (
	"github.com/Kora1128/FinSight/internal/broker/types"
	"github.com/Kora1128/FinSight/internal/broker/icici_direct"
	"github.com/Kora1128/FinSight/internal/broker/zerodha"
)

// ClientFactory creates broker clients
type ClientFactory interface {
	CreateZerodhaClient(apiKey, apiSecret string) types.Client
	CreateICICIDirectClient(apiKey, apiSecret string) types.Client
}

// DefaultClientFactory creates broker clients with the default implementations
type DefaultClientFactory struct{}

// CreateZerodhaClient creates a Zerodha client
func (f *DefaultClientFactory) CreateZerodhaClient(apiKey, apiSecret string) types.Client {
	return zerodha.NewClient(apiKey, apiSecret)
}

// CreateICICIDirectClient creates an ICICI Direct client
func (f *DefaultClientFactory) CreateICICIDirectClient(apiKey, apiSecret string) types.Client {
	return icici_direct.NewClient(apiKey, apiSecret)
}
