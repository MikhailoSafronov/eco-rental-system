package database

import (
	"context"
	"fmt"
	"math"
	"time"

	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartRide виконує безпечну транзакцію для початку оренди
func StartRide(pool *pgxpool.Pool, userID int, vehicleID int) (*models.Ride, error) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не вдалося почати транзакцію: %w", err)
	}
	defer tx.Rollback(ctx)

	var balance float64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1 AND deleted_at IS NULL", userID).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки балансу: %w", err)
	}
	if balance < 50.00 {
		return nil, fmt.Errorf("недостатньо коштів на балансі (мінімум 50.00 грн)")
	}

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM vehicles WHERE id = $1 FOR UPDATE", vehicleID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки самоката: %w", err)
	}
	if status != "available" {
		return nil, fmt.Errorf("самокат вже орендовано або він недоступний")
	}

	_, err = tx.Exec(ctx, "UPDATE vehicles SET status = 'rented', updated_at = NOW() WHERE id = $1", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("не вдалося оновити статус самоката: %w", err)
	}

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

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("не вдалося зафіксувати транзакцію: %w", err)
	}
	return ride, nil
}

// EndRide завершує активну поїздку користувача (НОВА ФУНКЦІЯ)
func EndRide(pool *pgxpool.Pool, userID int) (*models.Ride, error) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не вдалося почати транзакцію: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Шукаємо активну поїздку юзера та дістаємо тариф через JOIN
	queryInfo := `
		SELECT r.id, r.vehicle_id, r.start_time, t.unlock_price, t.minute_price
		FROM rides r
		JOIN vehicles v ON r.vehicle_id = v.id
		JOIN tariffs t ON v.tariff_id = t.id
		WHERE r.user_id = $1 AND r.status = 'active'
		FOR UPDATE OF r
	`
	var rideID, vehicleID int
	var startTime time.Time
	var unlockPrice, minutePrice float64

	err = tx.QueryRow(ctx, queryInfo, userID).Scan(&rideID, &vehicleID, &startTime, &unlockPrice, &minutePrice)
	if err != nil {
		return nil, fmt.Errorf("у вас немає активних поїздок")
	}

	// 2. Рахуємо час (округлюємо до хвилин у більшу сторону, мінімум 1 хвилина)
	duration := time.Since(startTime)
	minutes := math.Ceil(duration.Minutes())
	if minutes < 1 {
		minutes = 1 // Навіть якщо проїхав 10 секунд, платиш за 1 хвилину
	}
	totalPrice := unlockPrice + (minutes * minutePrice)

	// 3. Списуємо гроші з балансу користувача
	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance - $1 WHERE id = $2", totalPrice, userID)
	if err != nil {
		return nil, fmt.Errorf("помилка списання коштів: %w", err)
	}

	// 4. Повертаємо самокат на карту
	_, err = tx.Exec(ctx, "UPDATE vehicles SET status = 'available', updated_at = NOW() WHERE id = $1", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("не вдалося оновити статус самоката: %w", err)
	}

	// 5. Завершуємо поїздку в таблиці rides
	updateRideQuery := `
		UPDATE rides 
		SET status = 'completed', end_time = NOW(), total_price = $1 
		WHERE id = $2 
		RETURNING id, user_id, vehicle_id, status, start_time, end_time, total_price
	`
	ride := &models.Ride{}
	err = tx.QueryRow(ctx, updateRideQuery, totalPrice, rideID).Scan(
		&ride.ID, &ride.UserID, &ride.VehicleID, &ride.Status,
		&ride.StartTime, &ride.EndTime, &ride.TotalPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("помилка оновлення поїздки: %w", err)
	}

	// 6. Фіксуємо зміни
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("помилка збереження транзакції: %w", err)
	}

	return ride, nil
}
