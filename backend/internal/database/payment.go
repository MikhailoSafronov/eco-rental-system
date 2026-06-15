package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetUserPayments повертає історію транзакцій (списання та поповнення)
func GetUserPayments(pool *pgxpool.Pool, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, amount, type, status, created_at, external_transaction_id
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання платежів: %w", err)
	}
	defer rows.Close()

	var payments []map[string]interface{}
	for rows.Next() {
		var id int
		var amount float64
		var pType, status string
		var createdAt time.Time
		var extID *string

		if err := rows.Scan(&id, &amount, &pType, &status, &createdAt, &extID); err != nil {
			return nil, err
		}

		payments = append(payments, map[string]interface{}{
			"id":                      id,
			"amount":                  amount,
			"type":                    pType,
			"status":                  status,
			"created_at":              createdAt,
			"external_transaction_id": extID,
		})
	}
	return payments, nil
}
