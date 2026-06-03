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
	query := `
		UPDATE vehicles 
		SET 
			location = ST_SetSRID(ST_MakePoint($1, $2), 4326),
			battery_level = $3,
			updated_at = NOW()
		WHERE uuid = $4 AND deleted_at IS NULL
	`

	// ST_MakePoint приймає спочатку довготу (longitude), потім широту (latitude)
	commandTag, err := pool.Exec(context.Background(), query, lon, lat, battery, uuid)
	if err != nil {
		return fmt.Errorf("помилка оновлення телеметрії: %w", err)
	}

	// Перевіряємо, чи дійсно оновився хоча б один рядок (чи існує такий UUID)
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("самокат з UUID %s не знайдено", uuid)
	}

	return nil
}
