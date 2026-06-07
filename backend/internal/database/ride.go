package database

import (
	"context"
	"fmt"

	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartRide виконує безпечну транзакцію для початку оренди самоката
func StartRide(pool *pgxpool.Pool, userID int, vehicleID int) (*models.Ride, error) {
	ctx := context.Background()

	// 1. Починаємо транзакцію
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не вдалося почати транзакцію: %w", err)
	}

	// Якщо станеться помилка — скасовуємо всі часткові зміни
	defer tx.Rollback(ctx)

	// 2. Перевіряємо баланс користувача
	var balance float64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1 AND deleted_at IS NULL", userID).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки балансу: %w", err)
	}

	if balance < 50.00 {
		return nil, fmt.Errorf("недостатньо коштів на балансі (мінімум 50.00 грн)")
	}

	// 3. Перевіряємо статус самоката і блокуємо рядок від Race Condition
	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM vehicles WHERE id = $1 FOR UPDATE", vehicleID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки самоката: %w", err)
	}

	if status != "available" {
		return nil, fmt.Errorf("самокат вже орендовано або він недоступний")
	}

	// 4. Змінюємо статус самоката на 'rented'
	_, err = tx.Exec(ctx, "UPDATE vehicles SET status = 'rented', updated_at = NOW() WHERE id = $1", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("не вдалося оновити статус самоката: %w", err)
	}

	// 5. Створюємо запис про поїздку (копіюємо локацію з таблиці vehicles прямо в SQL)
	queryRide := `
		INSERT INTO rides (user_id, vehicle_id, status, start_time, start_location)
		VALUES ($1, $2, 'active', NOW(), (SELECT location FROM vehicles WHERE id = $2))
		RETURNING id, start_time
	`

	ride := &models.Ride{
		UserID:    userID,
		VehicleID: vehicleID,
		Status:    "active",
	}

	err = tx.QueryRow(ctx, queryRide, userID, vehicleID).Scan(&ride.ID, &ride.StartTime)
	if err != nil {
		return nil, fmt.Errorf("не вдалося створити поїздку: %w", err)
	}

	// 6. Фіксуємо зміни в базі даних
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("не вдалося зафіксувати транзакцію: %w", err)
	}

	return ride, nil
}
