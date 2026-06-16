package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AddVehicle додає новий транспортний засіб до бази даних
func AddVehicle(pool *pgxpool.Pool, modelID, tariffID int, lat, lon float64) (int, string, error) {
	query := `
		INSERT INTO vehicles (model_id, tariff_id, location, battery_level, status)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), 100, 'available')
		RETURNING id, uuid
	`
	var id int
	var uuid string
	// PostGIS очікує ST_MakePoint(longitude, latitude)
	err := pool.QueryRow(context.Background(), query, modelID, tariffID, lon, lat).Scan(&id, &uuid)
	if err != nil {
		return 0, "", fmt.Errorf("помилка додавання транспорту: %w", err)
	}
	return id, uuid, nil
}

// UpdateVehicleStatus змінює статус самоката (наприклад, відправляє на ремонт)
func UpdateVehicleStatus(pool *pgxpool.Pool, vehicleID int, newStatus string) error {
	query := `UPDATE vehicles SET status = $1, updated_at = NOW() WHERE id = $2`
	cmdTag, err := pool.Exec(context.Background(), query, newStatus, vehicleID)
	if err != nil {
		return fmt.Errorf("помилка оновлення статусу: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("самокат не знайдено")
	}
	return nil
}

// GetAllVehiclesAdmin повертає список абсолютно усіх самокатів для панелі адміністратора
func GetAllVehiclesAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			v.id, v.uuid, v.battery_level, v.status,
			ST_Y(v.location::geometry) AS latitude,
			ST_X(v.location::geometry) AS longitude,
			m.name AS model_name
		FROM vehicles v
		LEFT JOIN vehicle_models m ON v.model_id = m.id
		WHERE v.deleted_at IS NULL
		ORDER BY v.id ASC
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту до БД: %w", err)
	}
	defer rows.Close()

	// Ініціалізуємо пустим масивом, щоб уникнути помилок на фронтенді
	vehicles := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, battery int
		var uuid, status string
		var lat, lon float64
		var modelName *string

		if err := rows.Scan(&id, &uuid, &battery, &status, &lat, &lon, &modelName); err != nil {
			return nil, fmt.Errorf("помилка сканування рядка: %w", err)
		}

		mName := "Невідомо"
		if modelName != nil {
			mName = *modelName
		}

		vehicles = append(vehicles, map[string]interface{}{
			"id":            id,
			"uuid":          uuid,
			"battery_level": battery,
			"status":        status,
			"latitude":      lat,
			"longitude":     lon,
			"model_name":    mName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("помилка читання рядків транспорту: %w", err)
	}
	return vehicles, nil
}
