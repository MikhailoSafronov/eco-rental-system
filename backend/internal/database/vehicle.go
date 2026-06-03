package database

import (
	"context"
	"fmt"

	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAllAvailableVehicles повертає список усіх вільних самокатів
func GetAllAvailableVehicles(pool *pgxpool.Pool) ([]models.Vehicle, error) {
	// SQL-запит.
	// ST_Y і ST_X - це спеціальні функції PostGIS, які дістають координати з точки.
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

	// Створюємо порожній список (слайс) для самокатів
	var vehicles []models.Vehicle

	// Проходимось по кожному рядку, який повернула база
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

		// Додаємо знайдений самокат у наш список
		vehicles = append(vehicles, v)
	}

	return vehicles, nil
}
