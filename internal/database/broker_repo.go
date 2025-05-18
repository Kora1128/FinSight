package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/FinSight/internal/repository"
)

// Ensure BrokerCredentialsRepo implements repository.BrokerCredentialsRepository
var _ repository.BrokerCredentialsRepository = (*BrokerCredentialsRepo)(nil)

// BrokerCredentialsRepo handles broker credentials operations in the database
type BrokerCredentialsRepo struct {
	db *DB
}

// NewBrokerCredentialsRepo creates a new broker credentials repository
func NewBrokerCredentialsRepo(db *DB) *BrokerCredentialsRepo {
	return &BrokerCredentialsRepo{db: db}
}

// SaveCredentials saves broker credentials to the database
func (r *BrokerCredentialsRepo) SaveCredentials(userID string, brokerType string, apiKey string, apiSecret string) error {
	// Check if credentials already exist
	var id int
	err := r.db.QueryRow(
		"SELECT id FROM broker_credentials WHERE user_id = $1 AND broker_type = $2",
		userID, brokerType,
	).Scan(&id)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Insert new credentials
		_, err = r.db.Exec(
			"INSERT INTO broker_credentials (user_id, broker_type, api_key, api_secret) VALUES ($1, $2, $3, $4)",
			userID, brokerType, apiKey, apiSecret,
		)
	} else {
		// Update existing credentials
		_, err = r.db.Exec(
			"UPDATE broker_credentials SET api_key = $1, api_secret = $2, updated_at = $3 WHERE user_id = $4 AND broker_type = $5",
			apiKey, apiSecret, time.Now(), userID, brokerType,
		)
	}

	return err
}

// GetCredentials retrieves broker credentials from the database
func (r *BrokerCredentialsRepo) GetCredentials(userID string, brokerType string) (*models.Credentials, error) {
	credentials := &models.Credentials{}
	err := r.db.QueryRow(
		"SELECT id, user_id, broker_type, api_key, api_secret, access_token, token_expiry, created_at, updated_at FROM broker_credentials WHERE user_id = $1 AND broker_type = $2",
		userID, brokerType,
	).Scan(
		&credentials.ID,
		&credentials.UserID,
		&credentials.BrokerType,
		&credentials.APIKey,
		&credentials.APISecret,
		&credentials.AccessToken,
		&credentials.TokenExpiry,
		&credentials.CreatedAt,
		&credentials.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return credentials, nil
}

// UpdateAccessToken updates the access token and expiry time for broker credentials
func (r *BrokerCredentialsRepo) UpdateAccessToken(userID string, brokerType string, accessToken string, expiryTime time.Time) error {
	_, err := r.db.Exec(
		"UPDATE broker_credentials SET access_token = $1, token_expiry = $2, updated_at = $3 WHERE user_id = $4 AND broker_type = $5",
		accessToken, expiryTime, time.Now(), userID, brokerType,
	)
	return err
}

// DeleteCredentials deletes broker credentials from the database
func (r *BrokerCredentialsRepo) DeleteCredentials(userID string, brokerType string) error {
	_, err := r.db.Exec(
		"DELETE FROM broker_credentials WHERE user_id = $1 AND broker_type = $2",
		userID, brokerType,
	)
	return err
}

// HasCredentials checks if the user has credentials for a specific broker
func (r *BrokerCredentialsRepo) HasCredentials(userID string, brokerType string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM broker_credentials WHERE user_id = $1 AND broker_type = $2",
		userID, brokerType,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAccessToken retrieves the access token for a user's broker credentials
func (r *BrokerCredentialsRepo) GetAccessToken(userID string, brokerType string) (string, error) {
	var accessToken string
	err := r.db.QueryRow(
		"SELECT access_token FROM broker_credentials WHERE user_id = $1 AND broker_type = $2",
		userID, brokerType,
	).Scan(&accessToken)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return accessToken, nil
}

// GetCredentialsForAllUsers retrieves all broker credentials from the database
func (r *BrokerCredentialsRepo) GetCredentialsForAllUsers() ([]*models.Credentials, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, broker_type, api_key, api_secret, access_token, token_expiry, created_at, updated_at FROM broker_credentials",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*models.Credentials
	for rows.Next() {
		cred := &models.Credentials{}
		err := rows.Scan(
			&cred.ID,
			&cred.UserID,
			&cred.BrokerType,
			&cred.APIKey,
			&cred.APISecret,
			&cred.AccessToken,
			&cred.TokenExpiry,
			&cred.CreatedAt,
			&cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}

// GetExpiredTokens retrieves credentials with expired tokens
func (r *BrokerCredentialsRepo) GetExpiredTokens() ([]*models.Credentials, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, broker_type, api_key, api_secret, access_token, token_expiry, created_at, updated_at FROM broker_credentials WHERE token_expiry IS NOT NULL AND token_expiry < $1",
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*models.Credentials
	for rows.Next() {
		cred := &models.Credentials{}
		err := rows.Scan(
			&cred.ID,
			&cred.UserID,
			&cred.BrokerType,
			&cred.APIKey,
			&cred.APISecret,
			&cred.AccessToken,
			&cred.TokenExpiry,
			&cred.CreatedAt,
			&cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}
