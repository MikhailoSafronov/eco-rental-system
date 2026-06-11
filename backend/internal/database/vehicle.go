package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAllAvailableVehicles повертає список усіх вільних самокатів
func GetAllAvailableVehicles(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			v.id, 
			v.uuid, 
			v.battery_level, 
			v.status,
			ST_Y(v.location::geometry) AS latitude,
			ST_X(v.location::geometry) AS longitude,
			t.unlock_price,
			t.minute_price
		FROM vehicles v
		JOIN tariffs t ON v.tariff_id = t.id
		WHERE v.status = 'available' AND v.deleted_at IS NULL
	`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту до БД: %w", err)
	}
	defer rows.Close()

	vehicles := make([]map[string]interface{}, 0)

	for rows.Next() {
		var id, battery int
		var uuid, status string
		var lat, lon, unlockPrice, minutePrice float64

		if err := rows.Scan(&id, &uuid, &battery, &status, &lat, &lon, &unlockPrice, &minutePrice); err != nil {
			return nil, fmt.Errorf("помилка сканування рядка: %w", err)
		}

		vehicles = append(vehicles, map[string]interface{}{
			"id":            id,
			"uuid":          uuid,
			"battery_level": battery,
			"status":        status,
			"latitude":      lat,
			"longitude":     lon,
			"unlock_price":  unlockPrice,
			"minute_price":  minutePrice,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("помилка читання рядків транспорту: %w", err)
	}

	return vehicles, nil
}

// GetVehicleByID шукає конкретний транспортний засіб за його ID
func GetVehicleByID(pool *pgxpool.Pool, id int) (map[string]interface{}, error) {
	query := `
		SELECT 
			v.id, 
			v.uuid, 
			v.battery_level, 
			v.status,
			ST_Y(v.location::geometry) AS latitude,
			ST_X(v.location::geometry) AS longitude,
			t.unlock_price,
			t.minute_price
		FROM vehicles v
		JOIN tariffs t ON v.tariff_id = t.id
		WHERE v.id = $1 AND v.deleted_at IS NULL
	`

	var vID, battery int
	var uuid, status string
	var lat, lon, unlockPrice, minutePrice float64

	err := pool.QueryRow(context.Background(), query, id).Scan(&vID, &uuid, &battery, &status, &lat, &lon, &unlockPrice, &minutePrice)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            vID,
		"uuid":          uuid,
		"battery_level": battery,
		"status":        status,
		"latitude":      lat,
		"longitude":     lon,
		"unlock_price":  unlockPrice,
		"minute_price":  minutePrice,
	}, nil
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
