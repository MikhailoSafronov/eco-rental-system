package database

import (
	"context"
	"fmt"

	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAllAvailableVehicles повертає список усіх вільних самокатів
func GetAllAvailableVehicles(pool *pgxpool.Pool) ([]models.Vehicle, error) {
	query := `
		SELECT 
			id, 
			uuid, 
			battery_level, 
			status,
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude
		FROM vehicles
		WHERE status = 'available' AND deleted_at IS NULL
	`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту до БД: %w", err)
	}
	defer rows.Close()

	var vehicles []models.Vehicle

	for rows.Next() {
		var v models.Vehicle
		err := rows.Scan(
			&v.ID,
			&v.UUID,
			&v.BatteryLevel,
			&v.Status,
			&v.Latitude,
			&v.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("помилка сканування рядка: %w", err)
		}

		vehicles = append(vehicles, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("помилка читання рядків транспорту: %w", err)
	}

	return vehicles, nil
}

// GetVehicleByID шукає конкретний транспортний засіб за його ID
func GetVehicleByID(pool *pgxpool.Pool, id int) (*models.Vehicle, error) {
	query := `
		SELECT 
			id, 
			uuid, 
			battery_level, 
			status,
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude
		FROM vehicles
		WHERE id = $1 AND deleted_at IS NULL
	`

	var v models.Vehicle

	err := pool.QueryRow(context.Background(), query, id).Scan(
		&v.ID,
		&v.UUID,
		&v.BatteryLevel,
		&v.Status,
		&v.Latitude,
		&v.Longitude,
	)

	if err != nil {
		return nil, err
	}

	return &v, nil
}

// UpdateVehicleTelemetry оновлює координати та заряд батареї самоката за його UUID
func UpdateVehicleTelemetry(pool *pgxpool.Pool, uuid string, lat float64, lon float64, battery int) error {
	ctx := context.Background()

	// Відкриваємо транзакцію, оскільки тепер у нас два пов'язаних запити
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("помилка старту транзакції: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Оновлюємо сам самокат і одразу дістаємо його ID та поточний статус (RETURNING)
	updateVehicleQuery := `
		UPDATE vehicles 
		SET 
			location = ST_SetSRID(ST_MakePoint($1, $2), 4326),
			battery_level = $3,
			updated_at = NOW()
		WHERE uuid = $4 AND deleted_at IS NULL
		RETURNING id, status
	`

	var vehicleID int
	var status string

	err = tx.QueryRow(ctx, updateVehicleQuery, lon, lat, battery, uuid).Scan(&vehicleID, &status)
	if err != nil {
		// Якщо QueryRow повертає помилку, що рядків не знайдено
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("самокат з UUID %s не знайдено", uuid)
		}
		return fmt.Errorf("помилка оновлення телеметрії: %w", err)
	}

	// 2. Якщо самокат зараз в оренді, записуємо точку маршруту
	if status == "rented" {
		insertTelemetryQuery := `
			INSERT INTO ride_telemetry (ride_id, location)
			SELECT id, ST_SetSRID(ST_MakePoint($1, $2), 4326)
			FROM rides
			WHERE vehicle_id = $3 AND status = 'active'
		`
		_, err = tx.Exec(ctx, insertTelemetryQuery, lon, lat, vehicleID)
		if err != nil {
			return fmt.Errorf("помилка збереження треку поїздки: %w", err)
		}
	}

	// 3. Фіксуємо зміни
	return tx.Commit(ctx)
}
