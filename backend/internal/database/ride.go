package database

import (
	"context"
	"encoding/json"
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

	// ПЕРЕВІРКА: Чи є у користувача вже активна поїздка?
	var activeRides int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM rides WHERE user_id = $1 AND status = 'active'", userID).Scan(&activeRides)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки активних поїздок: %w", err)
	}
	if activeRides > 0 {
		return nil, fmt.Errorf("у вас вже є активна поїздка. Одночасно можна орендувати лише один транспортний засіб")
	}

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM vehicles WHERE id = $1 FOR UPDATE", vehicleID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки транспорту: %w", err)
	}
	if status != "available" {
		return nil, fmt.Errorf("транспорт вже орендовано або він недоступний")
	}

	_, err = tx.Exec(ctx, "UPDATE vehicles SET status = 'rented', updated_at = NOW() WHERE id = $1", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("не вдалося оновити статус транспорту: %w", err)
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
func EndRide(pool *pgxpool.Pool, userID int, photoURL string) (*models.Ride, error) {
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

	// =================================================================
	// КРОК 1.5: ПЕРЕВІРКА ГЕОЗОНИ ПАРКУВАННЯ (POSTGIS)
	// =================================================================
	var isParkedCorrectly bool
	zoneCheckQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM parking_zones pz
			INNER JOIN vehicles v ON v.id = $1
			WHERE pz.is_active = true
			AND ST_Intersects(pz.polygon, v.location)
		)
	`
	err = tx.QueryRow(ctx, zoneCheckQuery, vehicleID).Scan(&isParkedCorrectly)
	if err != nil {
		return nil, fmt.Errorf("помилка перевірки паркувальних зон: %w", err)
	}
	if !isParkedCorrectly {
		return nil, fmt.Errorf("Завершення поїздки неможливе! Транспорт знаходиться поза межами дозволеної зони паркування (Центр Херсона). Будь ласка, перепаркуйте транспорт.")
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

	// 4. Повертаємо транспорт на карту
	_, err = tx.Exec(ctx, "UPDATE vehicles SET status = 'available', updated_at = NOW() WHERE id = $1", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("не вдалося оновити статус транспорту: %w", err)
	}

	// 5. Завершуємо поїздку в таблиці rides
	updateRideQuery := `
		UPDATE rides 
		SET status = 'completed', end_time = NOW(), total_price = $1, end_photo_url = $2
		WHERE id = $3 
		RETURNING id, user_id, vehicle_id, status, start_time, end_time, total_price
	`
	ride := &models.Ride{}
	err = tx.QueryRow(ctx, updateRideQuery, totalPrice, photoURL, rideID).Scan(
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

// GetUserRideHistory повертає історію поїздок користувача
func GetUserRideHistory(pool *pgxpool.Pool, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			r.id, r.vehicle_id, r.status, r.start_time, r.end_time, r.total_price, r.end_photo_url,
			COALESCE(
				(SELECT json_agg(
					json_build_object(
						'latitude', ST_Y(location::geometry),
						'longitude', ST_X(location::geometry),
						'timestamp', timestamp
					) ORDER BY timestamp ASC
				) FROM ride_telemetry rt WHERE rt.ride_id = r.id),
				'[]'::json
			) AS track
		FROM rides r
		WHERE r.user_id = $1
		ORDER BY r.start_time DESC
	`
	rows, err := pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("помилка отримання історії: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id, vehicleID int
		var status string
		var startTime time.Time
		var endTime *time.Time
		var totalPrice float64
		var endPhotoURL *string
		var trackJSON []byte

		if err := rows.Scan(&id, &vehicleID, &status, &startTime, &endTime, &totalPrice, &endPhotoURL, &trackJSON); err != nil {
			return nil, err
		}

		var track []map[string]interface{}
		if err := json.Unmarshal(trackJSON, &track); err != nil {
			return nil, fmt.Errorf("помилка парсингу треку: %w", err)
		}

		history = append(history, map[string]interface{}{
			"id":            id,
			"vehicle_id":    vehicleID,
			"status":        status,
			"start_time":    startTime,
			"end_time":      endTime,
			"total_price":   totalPrice,
			"end_photo_url": endPhotoURL,
			"track":         track,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("помилка читання рядків: %w", err)
	}

	if history == nil {
		history = []map[string]interface{}{} // Повертаємо пустий масив, якщо поїздок ще немає
	}
	return history, nil
}
