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

	// Використовуємо QueryRow, бо шукаємо лише один запис
	err := pool.QueryRow(context.Background(), query, id).Scan(
		&v.ID,
		&v.UUID,
		&v.BatteryLevel,
		&v.Status,
		&v.Latitude,
		&v.Longitude,
	)

	if err != nil {
		return nil, err // Повертаємо помилку (наприклад, якщо самокат не знайдено)
	}

	return &v, nil
}
