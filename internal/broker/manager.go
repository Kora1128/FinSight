package broker

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kora1128/FinSight/internal/broker/types"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/repository"
)

// ClientType represents the type of broker client
type ClientType string

const (
	// ClientTypeZerodha represents a Zerodha client
	ClientTypeZerodha ClientType = "zerodha"
	// ClientTypeICICIDirect represents an ICICI Direct client
	ClientTypeICICIDirect ClientType = "icici_direct"
)

// ClientCredentials represents the credentials needed to create a broker client
type ClientCredentials struct {
	APIKey       string
	APISecret    string
	RequestToken string
	Password     string // For ICICI Direct
	UserID       string // Unique user identifier
}

// ClientRegistry holds user-specific broker clients
type ClientRegistry struct {
	userID        string
	zerodhaClient types.Client
	iciciClient   types.Client
	createdAt     time.Time
	lastAccessed  time.Time
	lastRefreshed time.Time
}

// BrokerManager manages the creation and refreshing of broker clients
type BrokerManager struct {
	credentialsRepo repository.BrokerCredentialsRepository
	cache           *cache.Cache // Cache only for tokens that need quick access
	registry        map[string]*ClientRegistry // Map from userID to ClientRegistry
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
		registry:        make(map[string]*ClientRegistry),
		maxAge:          maxAge,
		refreshInterval: refreshInterval,
		factory:         &DefaultClientFactory{},
	}
	
	// Start a background goroutine to periodically refresh tokens
	go manager.startRefreshWorker()
	
	return manager
}

// GetOrCreateClient gets an existing client or creates a new one based on the provided credentials
func (m *BrokerManager) GetOrCreateClient(clientType ClientType, creds ClientCredentials) (types.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if we already have a registry for this user
	registry, exists := m.registry[creds.UserID]
	if !exists {
		// Create a new registry for this user
		registry = &ClientRegistry{
			userID:       creds.UserID,
			createdAt:    time.Now(),
			lastAccessed: time.Now(),
		}
		m.registry[creds.UserID] = registry
	}
	
	// Update last accessed time
	registry.lastAccessed = time.Now()
	
	switch clientType {
	case ClientTypeZerodha:
		if registry.zerodhaClient != nil {
			return registry.zerodhaClient, nil
		}
		
		// Create a new Zerodha client
		client := m.factory.CreateZerodhaClient(creds.APIKey, creds.APISecret)
		
		// Authenticate if request token is provided
		if creds.RequestToken != "" {
			if err := client.Login(creds.RequestToken, creds.APISecret); err != nil {
				return nil, fmt.Errorf("failed to login to Zerodha: %w", err)
			}
			
			// Store the token in database
			err := m.credentialsRepo.UpdateAccessToken(creds.UserID, string(ClientTypeZerodha), client.GetAccessToken(), time.Now().Add(m.maxAge))
			if err != nil {
				return nil, fmt.Errorf("failed to update access token in database: %w", err)
			}
			
			// Also keep in cache for quick access
			tokenKey := fmt.Sprintf("zerodha_token:%s", creds.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}
		
		registry.zerodhaClient = client
		return client, nil
		
	case ClientTypeICICIDirect:
		if registry.iciciClient != nil {
			return registry.iciciClient, nil
		}
		
		// Create a new ICICI Direct client
		client := m.factory.CreateICICIDirectClient(creds.APIKey, creds.APISecret)
		
		// Authenticate if request token is provided
		if creds.RequestToken != "" {
			if err := client.Login(creds.RequestToken, creds.APISecret); err != nil {
				return nil, fmt.Errorf("failed to login to ICICI Direct: %w", err)
			}
			
			// Store the token in database
			err := m.credentialsRepo.UpdateAccessToken(creds.UserID, string(ClientTypeICICIDirect), client.GetAccessToken(), time.Now().Add(m.maxAge))
			if err != nil {
				return nil, fmt.Errorf("failed to update access token in database: %w", err)
			}
			
			// Also keep in cache for quick access
			tokenKey := fmt.Sprintf("icici_token:%s", creds.UserID)
			m.cache.Set(tokenKey, client.GetAccessToken(), m.maxAge)
		}
		
		registry.iciciClient = client
		return client, nil
		
	default:
		return nil, fmt.Errorf("unknown client type: %s", clientType)
	}
}

// GetClient returns a client for the given user and client type, if it exists
func (m *BrokerManager) GetClient(userID string, clientType ClientType) (types.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	registry, exists := m.registry[userID]
	if !exists {
		return nil, false
	}

	registry.lastAccessed = time.Now()

	switch clientType {
	case ClientTypeZerodha:
		return registry.zerodhaClient, registry.zerodhaClient != nil
	case ClientTypeICICIDirect:
		return registry.iciciClient, registry.iciciClient != nil
	default:
		return nil, false
	}
}

// RefreshTokens attempts to refresh the tokens for all clients of a specific user
func (m *BrokerManager) RefreshTokens(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	registry, exists := m.registry[userID]
	if !exists {
		return fmt.Errorf("no clients found for user: %s", userID)
	}

	// Refresh Zerodha token if client exists
	if registry.zerodhaClient != nil && registry.zerodhaClient.CanAutoRefresh() {
		if err := registry.zerodhaClient.RefreshToken(); err != nil {
			return fmt.Errorf("failed to refresh Zerodha token: %w", err)
		}

		// Update token in database
		err := m.credentialsRepo.UpdateAccessToken(userID, string(ClientTypeZerodha), registry.zerodhaClient.GetAccessToken(), time.Now().Add(m.maxAge))
		if err != nil {
			return fmt.Errorf("failed to update Zerodha access token in database: %w", err)
		}

		// Also keep in cache for quick access
		tokenKey := fmt.Sprintf("zerodha_token:%s", userID)
		m.cache.Set(tokenKey, registry.zerodhaClient.GetAccessToken(), m.maxAge)
	}

	// Refresh ICICI Direct token if client exists
	if registry.iciciClient != nil && registry.iciciClient.CanAutoRefresh() {
		if err := registry.iciciClient.RefreshToken(); err != nil {
			return fmt.Errorf("failed to refresh ICICI Direct token: %w", err)
		}

		// Update token in database
		err := m.credentialsRepo.UpdateAccessToken(userID, string(ClientTypeICICIDirect), registry.iciciClient.GetAccessToken(), time.Now().Add(m.maxAge))
		if err != nil {
			return fmt.Errorf("failed to update ICICI Direct access token in database: %w", err)
		}

		// Also keep in cache for quick access
		tokenKey := fmt.Sprintf("icici_token:%s", userID)
		m.cache.Set(tokenKey, registry.iciciClient.GetAccessToken(), m.maxAge)
	}

	registry.lastRefreshed = time.Now()
	return nil
}

// startRefreshWorker starts a goroutine that periodically refreshes tokens
func (m *BrokerManager) startRefreshWorker() {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.refreshAllTokens()
		m.cleanupStaleClients()
	}
}

// refreshAllTokens attempts to refresh tokens for all clients
func (m *BrokerManager) refreshAllTokens() {
	m.mu.RLock()
	userIDs := make([]string, 0, len(m.registry))
	for userID := range m.registry {
		userIDs = append(userIDs, userID)
	}
	m.mu.RUnlock()

	for _, userID := range userIDs {
		// Ignore errors, as some clients may not have refresh tokens
		_ = m.RefreshTokens(userID)
	}
}

// cleanupStaleClients removes clients that haven't been accessed for a long time
func (m *BrokerManager) cleanupStaleClients() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	staleThreshold := now.Add(-m.maxAge)

	for userID, registry := range m.registry {
		if registry.lastAccessed.Before(staleThreshold) {
			delete(m.registry, userID)
		}
	}
}

// RemoveClient removes a client for a specific user and client type
func (m *BrokerManager) RemoveClient(userID string, clientType ClientType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	registry, exists := m.registry[userID]
	if !exists {
		return
	}

	switch clientType {
	case ClientTypeZerodha:
		registry.zerodhaClient = nil
		// Remove from cache
		tokenKey := fmt.Sprintf("zerodha_token:%s", userID)
		m.cache.Delete(tokenKey)

		// Remove from database
		_ = m.credentialsRepo.DeleteCredentials(userID, string(ClientTypeZerodha))

	case ClientTypeICICIDirect:
		registry.iciciClient = nil
		// Remove from cache
		tokenKey := fmt.Sprintf("icici_token:%s", userID)
		m.cache.Delete(tokenKey)

		refreshTokenKey := fmt.Sprintf("icici_refresh_token:%s", userID)
		m.cache.Delete(refreshTokenKey)

		// Remove from database
		_ = m.credentialsRepo.DeleteCredentials(userID, string(ClientTypeICICIDirect))
	}

	// If both clients are nil, remove the registry entirely
	if registry.zerodhaClient == nil && registry.iciciClient == nil {
		delete(m.registry, userID)
	}
}
