package database

import (
	"time"

	"github.com/Kora1128/FinSight/internal/models"
)

// PortfolioRepo handles portfolio operations in the database
type PortfolioRepo struct {
	db *DB
}

// NewPortfolioRepo creates a new portfolio repository
func NewPortfolioRepo(db *DB) *PortfolioRepo {
	return &PortfolioRepo{db: db}
}

// SaveHoldings saves portfolio holdings to the database for a user
func (r *PortfolioRepo) SaveHoldings(userID string, holdings []models.Holding) error {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete existing holdings
	_, err = tx.Exec("DELETE FROM portfolio_holdings WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// Insert new holdings
	for _, holding := range holdings {
		_, err = tx.Exec(
			`INSERT INTO portfolio_holdings 
			(user_id, item_name, isin, quantity, average_price, last_traded_price, 
			current_value, day_change, day_change_percent, total_pnl, platform, holding_type, last_updated) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			userID,
			holding.ItemName,
			holding.ISIN,
			holding.Quantity,
			holding.AveragePrice,
			holding.LastTradedPrice,
			holding.CurrentValue,
			holding.DayChange,
			holding.DayChangePercent,
			holding.TotalPnL,
			holding.Platform,
			holding.Type,
			holding.LastUpdated,
		)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// GetHoldings retrieves portfolio holdings for a user
func (r *PortfolioRepo) GetHoldings(userID string) ([]models.Holding, error) {
	rows, err := r.db.Query(
		`SELECT item_name, isin, quantity, average_price, last_traded_price, 
		current_value, day_change, day_change_percent, total_pnl, platform, holding_type, last_updated 
		FROM portfolio_holdings WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		var holding models.Holding
		err := rows.Scan(
			&holding.ItemName,
			&holding.ISIN,
			&holding.Quantity,
			&holding.AveragePrice,
			&holding.LastTradedPrice,
			&holding.CurrentValue,
			&holding.DayChange,
			&holding.DayChangePercent,
			&holding.TotalPnL,
			&holding.Platform,
			&holding.Type,
			&holding.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return holdings, nil
}

// GetPlatformHoldings retrieves portfolio holdings for a user filtered by platform
func (r *PortfolioRepo) GetPlatformHoldings(userID string, platform string) ([]models.Holding, error) {
	rows, err := r.db.Query(
		`SELECT item_name, isin, quantity, average_price, last_traded_price, 
		current_value, day_change, day_change_percent, total_pnl, platform, holding_type, last_updated 
		FROM portfolio_holdings WHERE user_id = $1 AND platform = $2`,
		userID, platform,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		var holding models.Holding
		err := rows.Scan(
			&holding.ItemName,
			&holding.ISIN,
			&holding.Quantity,
			&holding.AveragePrice,
			&holding.LastTradedPrice,
			&holding.CurrentValue,
			&holding.DayChange,
			&holding.DayChangePercent,
			&holding.TotalPnL,
			&holding.Platform,
			&holding.Type,
			&holding.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return holdings, nil
}

// GetHoldingsByType retrieves portfolio holdings for a user filtered by holding type
func (r *PortfolioRepo) GetHoldingsByType(userID string, holdingType models.HoldingType) ([]models.Holding, error) {
	rows, err := r.db.Query(
		`SELECT item_name, isin, quantity, average_price, last_traded_price, 
		current_value, day_change, day_change_percent, total_pnl, platform, holding_type, last_updated 
		FROM portfolio_holdings WHERE user_id = $1 AND holding_type = $2`,
		userID, holdingType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		var holding models.Holding
		err := rows.Scan(
			&holding.ItemName,
			&holding.ISIN,
			&holding.Quantity,
			&holding.AveragePrice,
			&holding.LastTradedPrice,
			&holding.CurrentValue,
			&holding.DayChange,
			&holding.DayChangePercent,
			&holding.TotalPnL,
			&holding.Platform,
			&holding.Type,
			&holding.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return holdings, nil
}

// GetPortfolioLastUpdated gets the timestamp when the portfolio was last updated
func (r *PortfolioRepo) GetPortfolioLastUpdated(userID string) (time.Time, bool, error) {
	var lastUpdated time.Time
	err := r.db.QueryRow(
		"SELECT MAX(last_updated) FROM portfolio_holdings WHERE user_id = $1",
		userID,
	).Scan(&lastUpdated)

	if err != nil {
		return time.Time{}, false, err
	}

	if lastUpdated.IsZero() {
		return time.Time{}, false, nil
	}

	return lastUpdated, true, nil
}
