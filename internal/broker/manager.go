package broker

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kora1128/FinSight/internal/broker/types"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/FinSight/internal/repository"
)

const (
	// ClientTypeZerodha represents a Zerodha client
	ClientTypeZerodha string = "zerodha"
	// ClientTypeICICIDirect represents an ICICI Direct client
	ClientTypeICICIDirect string = "icici_direct"
)

// ClientCredentials represents the credentials needed to create a broker client
type ClientCredentials struct {
	APIKey       string
	APISecret    string
	RequestToken string
	Password     string // For ICICI Direct
	UserID       string // Unique user identifier
}

// BrokerManager manages the creation and refreshing of broker clients
type BrokerManager struct {
	credentialsRepo repository.BrokerCredentialsRepository
	cache           *cache.Cache // Cache only for tokens that need quick access
	mu              sync.RWMutex
	maxAge          time.Duration // Maximum age before a client is considered stale
	refreshInterval time.Duration
	factory         ClientFactory
}

// NewBrokerManager creates a new BrokerManager
func NewBrokerManager(credentialsRepo repository.BrokerCredentialsRepository, cache *cache.Cache, maxAge, refreshInterval time.Duration) *BrokerManager {
	manager := &BrokerManager{
		credentialsRepo: credentialsRepo,
		cache:           cache,
		maxAge:          maxAge,
		refreshInterval: refreshInterval,
		factory:         &DefaultClientFactory{},
	}

	// Start a background goroutine to periodically refresh tokens
	go manager.startRefreshWorker()

	return manager
}

// CreateClient gets an existing client or creates a new one based on the provided credentials
func (m *BrokerManager) CreateClient(clientType string, creds ClientCredentials) (types.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch clientType {
	case ClientTypeZerodha:

		// Create a new Zerodha client
		client := m.factory.CreateZerodhaClient(creds.APIKey, creds.APISecret, creds.RequestToken)

		// Authenticate if request token is provided
		if creds.RequestToken != "" {
			if err := client.Login(); err != nil {
				return nil, fmt.Errorf("failed to login to Zerodha: %w", err)
			}

			// Store the token in database
			err := m.credentialsRepo.SaveCredentials(creds.UserID, string(ClientTypeZerodha), creds.APIKey, creds.APISecret, creds.RequestToken, time.Now().Add(m.maxAge))
			if err != nil {
				return nil, fmt.Errorf("failed to update access token in database: %w", err)
			}

			// Also keep in cache for quick access

			tokenKey := fmt.Sprintf("%s:%s", cache.KeyZerodhaToken, creds.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}

		return client, nil

	case ClientTypeICICIDirect:

		// Create a new ICICI Direct client
		client := m.factory.CreateICICIDirectClient(creds.APIKey, creds.APISecret, creds.RequestToken)

		// Authenticate if request token is provided
		if creds.RequestToken != "" {
			if err := client.Login(); err != nil {
				return nil, fmt.Errorf("failed to login to ICICI Direct: %w", err)
			}

			// Store the token in database
			err := m.credentialsRepo.SaveCredentials(creds.UserID, string(ClientTypeICICIDirect), creds.APIKey, creds.APISecret, creds.RequestToken, time.Now().Add(m.maxAge))
			if err != nil {
				return nil, fmt.Errorf("failed to update access token in database: %w", err)
			}

			// Also keep in cache for quick access
			tokenKey := fmt.Sprintf("%s:%s", cache.KeyICICIToken, creds.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}
		return client, nil

	default:
		return nil, fmt.Errorf("unknown client type: %s", clientType)
	}
}

// GetClient returns a client for the given user and client type, if it exists
func (m *BrokerManager) GetClient(userID string, clientType string) (types.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	creds, err := m.credentialsRepo.GetCredentials(userID, string(clientType))
	if err != nil {
		return nil, false
	}

	switch clientType {
	case ClientTypeZerodha:
		client := m.factory.CreateZerodhaClient(creds.APIKey, creds.APISecret, creds.AccessToken)
		tokenKey := fmt.Sprintf("%s:%s", cache.KeyZerodhaToken, creds.UserID)
		if token, found := m.cache.Get(tokenKey); found {
			client.SetAccessToken(token.(string))
		} else {
			return nil, false
		}
		return client, true
	case ClientTypeICICIDirect:
		client := m.factory.CreateICICIDirectClient(creds.APIKey, creds.APISecret, creds.AccessToken)
		tokenKey := fmt.Sprintf("%s:%s", cache.KeyICICIToken, creds.UserID)
		if token, found := m.cache.Get(tokenKey); found {
			client.SetAccessToken(token.(string))
		} else {
			return nil, false
		}
		return client, true
	default:
		return nil, false
	}

}

// RefreshTokens attempts to refresh the tokens for all clients of a specific user
func (m *BrokerManager) RefreshTokens(userID string, cred *models.Credentials) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch cred.BrokerType {
	case ClientTypeZerodha:

		// Create a new Zerodha client
		client := m.factory.CreateZerodhaClient(cred.APIKey, cred.APISecret, cred.AccessToken)

		// Authenticate if request token is provided
		if cred.AccessToken != "" {
			if err := client.Login(); err != nil {
				return fmt.Errorf("failed to login to Zerodha: %w", err)
			}

			// Also keep in cache for quick access

			tokenKey := fmt.Sprintf("%s:%s", cache.KeyZerodhaToken, cred.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}

		return nil

	case ClientTypeICICIDirect:

		// Create a new ICICI Direct client
		client := m.factory.CreateICICIDirectClient(cred.APIKey, cred.APISecret, cred.AccessToken)

		// Authenticate if request token is provided
		if cred.AccessToken != "" {
			if err := client.Login(); err != nil {
				return fmt.Errorf("failed to login to ICICI Direct: %w", err)
			}

			// Also keep in cache for quick access
			tokenKey := fmt.Sprintf("%s:%s", cache.KeyICICIToken, cred.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}
		return nil

	default:
		return fmt.Errorf("unknown client type: %s", cred.BrokerType)
	}
}

// startRefreshWorker starts a goroutine that periodically refreshes tokens
func (m *BrokerManager) startRefreshWorker() {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupStaleClients()
		m.refreshAllTokens()
	}
}

// refreshAllTokens attempts to refresh tokens for all clients
func (m *BrokerManager) refreshAllTokens() {
	m.mu.RLock()
	creds, err := m.credentialsRepo.GetCredentialsForAllUsers()
	if err != nil {
		return
	}
	for _, cred := range creds {
		userID := cred.UserID
		if _, found := m.cache.Get(fmt.Sprintf("%s:%s", cred.BrokerType+"_token", userID)); !found {
			_ = m.RefreshTokens(userID, cred)
		}
	}
	m.mu.RUnlock()
}

// cleanupStaleClients removes clients that haven't been accessed for a long time
func (m *BrokerManager) cleanupStaleClients() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache.DeleteExpired()
}

// RemoveClient removes a client for a specific user and client type
func (m *BrokerManager) RemoveClient(userID string, clientType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch clientType {
	case ClientTypeZerodha:
		// Remove from cache
		tokenKey := fmt.Sprintf("%s:%s", cache.KeyZerodhaToken, userID)
		m.cache.Delete(tokenKey)

		// Remove from database
		_ = m.credentialsRepo.DeleteCredentials(userID, string(ClientTypeZerodha))

	case ClientTypeICICIDirect:
		// Remove from cache
		tokenKey := fmt.Sprintf("%s:%s", cache.KeyICICIToken, userID)
		m.cache.Delete(tokenKey)

		// Remove from database
		_ = m.credentialsRepo.DeleteCredentials(userID, string(ClientTypeICICIDirect))
	}
}
