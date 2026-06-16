package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AddVehicle додає новий транспортний засіб до бази даних
func AddVehicle(pool *pgxpool.Pool, modelID, tariffID, batteryLevel int, lat, lon float64) (int, string, error) {
	query := `
		INSERT INTO vehicles (model_id, tariff_id, location, battery_level, status)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, 'available')
		RETURNING id, uuid
	`
	var id int
	var uuid string
	// PostGIS очікує ST_MakePoint(longitude, latitude)
	err := pool.QueryRow(context.Background(), query, modelID, tariffID, lon, lat, batteryLevel).Scan(&id, &uuid)
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
			m.name AS model_name,
			m.type AS vehicle_type
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
		var vehicleType *string

		if err := rows.Scan(&id, &uuid, &battery, &status, &lat, &lon, &modelName, &vehicleType); err != nil {
			return nil, fmt.Errorf("помилка сканування рядка: %w", err)
		}

		mName := "Невідомо"
		if modelName != nil {
			mName = *modelName
		}
		vType := "unknown"
		if vehicleType != nil {
			vType = *vehicleType
		}

		vehicles = append(vehicles, map[string]interface{}{
			"id":            id,
			"uuid":          uuid,
			"battery_level": battery,
			"status":        status,
			"latitude":      lat,
			"longitude":     lon,
			"model_name":    mName,
			"vehicle_type":  vType,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("помилка читання рядків транспорту: %w", err)
	}
	return vehicles, nil
}

// GetAllModelsAdmin повертає всі моделі транспорту з бази (для випадаючого списку в адмінці)
func GetAllModelsAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `SELECT id, name, type FROM vehicle_models WHERE deleted_at IS NULL ORDER BY id ASC`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту моделей: %w", err)
	}
	defer rows.Close()

	models := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var name, vType string
		if err := rows.Scan(&id, &name, &vType); err != nil {
			return nil, err
		}
		models = append(models, map[string]interface{}{
			"id":   id,
			"name": name,
			"type": vType,
		})
	}
	return models, nil
}

// GetAllTariffsAdmin повертає всі тарифи з бази (для випадаючого списку в адмінці)
func GetAllTariffsAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `SELECT id, name, unlock_price, minute_price FROM tariffs ORDER BY id ASC`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту тарифів: %w", err)
	}
	defer rows.Close()

	tariffs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var name string
		var unlock, minute float64
		if err := rows.Scan(&id, &name, &unlock, &minute); err != nil {
			return nil, err
		}
		tariffs = append(tariffs, map[string]interface{}{
			"id":           id,
			"name":         name,
			"unlock_price": unlock,
			"minute_price": minute,
		})
	}
	return tariffs, nil
}

// AddParkingZone додає нову паркувальну зону з вказаним полігоном
func AddParkingZone(pool *pgxpool.Pool, name string, wktPolygon string) (int, error) {
	query := `INSERT INTO parking_zones (name, polygon) VALUES ($1, ST_GeomFromText($2, 4326)) RETURNING id`
	var id int
	if err := pool.QueryRow(context.Background(), query, name, wktPolygon).Scan(&id); err != nil {
		return 0, fmt.Errorf("помилка збереження зони: %w", err)
	}
	return id, nil
}
