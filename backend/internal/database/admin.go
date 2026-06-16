package database

import (
	"context"
	"fmt"
	"time"

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

// GetAllRidesAdmin повертає історію всіх поїздок для адмін-панелі
func GetAllRidesAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			r.id, r.user_id, u.email AS user_email, 
			r.vehicle_id, v.uuid AS vehicle_uuid, 
			r.status, r.start_time, r.end_time, r.total_price
		FROM rides r
		JOIN users u ON r.user_id = u.id
		JOIN vehicles v ON r.vehicle_id = v.id
		ORDER BY r.start_time DESC
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту поїздок: %w", err)
	}
	defer rows.Close()

	rides := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, userID, vehicleID int
		var userEmail, vehicleUUID, status string
		var startTime time.Time
		var endTime *time.Time
		var totalPrice float64

		if err := rows.Scan(&id, &userID, &userEmail, &vehicleID, &vehicleUUID, &status, &startTime, &endTime, &totalPrice); err != nil {
			return nil, err
		}
		rides = append(rides, map[string]interface{}{
			"id":           id,
			"user_id":      userID,
			"user_email":   userEmail,
			"vehicle_id":   vehicleID,
			"vehicle_uuid": vehicleUUID,
			"status":       status,
			"start_time":   startTime,
			"end_time":     endTime,
			"total_price":  totalPrice,
		})
	}
	return rides, nil
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
	query := `SELECT id, name, type, battery_capacity_wh, max_speed FROM vehicle_models WHERE deleted_at IS NULL ORDER BY type, id ASC`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту моделей: %w", err)
	}
	defer rows.Close()

	models := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, battery, speed int
		var name, vType string
		if err := rows.Scan(&id, &name, &vType, &battery, &speed); err != nil {
			return nil, err
		}
		models = append(models, map[string]interface{}{
			"id":                  id,
			"name":                name,
			"type":                vType,
			"battery_capacity_wh": battery,
			"max_speed":           speed,
		})
	}
	return models, nil
}

// GetAllTariffsAdmin повертає всі тарифи з бази (для випадаючого списку в адмінці)
func GetAllTariffsAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `SELECT id, name, type, unlock_price, minute_price FROM tariffs ORDER BY type, id ASC`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту тарифів: %w", err)
	}
	defer rows.Close()

	tariffs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var name, vType string
		var unlock, minute float64
		if err := rows.Scan(&id, &name, &vType, &unlock, &minute); err != nil {
			return nil, err
		}
		tariffs = append(tariffs, map[string]interface{}{
			"id":           id,
			"name":         name,
			"vehicle_type": vType,
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

// DeleteVehicle виконує "м'яке" видалення (Soft Delete) транспорту
func DeleteVehicle(pool *pgxpool.Pool, id int) error {
	// Ми не видаляємо рядок фізично, а лише ставимо мітку deleted_at.
	// Додатково переводимо статус у 'maintenance', щоб ніхто не міг його орендувати.
	query := `UPDATE vehicles SET deleted_at = NOW(), status = 'maintenance' WHERE id = $1`
	cmdTag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("помилка видалення транспорту: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("транспорт не знайдено")
	}
	return nil
}

// AddVehicleModel додає нову модель транспорту
func AddVehicleModel(pool *pgxpool.Pool, name, vType string, batteryCapacity, maxSpeed int) (int, error) {
	query := `INSERT INTO vehicle_models (name, type, battery_capacity_wh, max_speed) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	if err := pool.QueryRow(context.Background(), query, name, vType, batteryCapacity, maxSpeed).Scan(&id); err != nil {
		return 0, fmt.Errorf("помилка створення моделі: %w", err)
	}
	return id, nil
}

// DeleteVehicleModel виконує "м'яке" видалення моделі (щоб не зламати існуючі самокати)
func DeleteVehicleModel(pool *pgxpool.Pool, id int) error {
	query := `UPDATE vehicle_models SET deleted_at = NOW() WHERE id = $1`
	cmdTag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("помилка видалення моделі: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("модель не знайдено")
	}
	return nil
}

// DeleteParkingZone видаляє зону паркування з бази
func DeleteParkingZone(pool *pgxpool.Pool, id int) error {
	query := `DELETE FROM parking_zones WHERE id = $1`
	cmdTag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("помилка видалення зони: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("зону не знайдено")
	}
	return nil
}

// UpdateTariff оновлює ціни існуючого тарифу
func UpdateTariff(pool *pgxpool.Pool, id int, unlockPrice, minutePrice float64) error {
	query := `UPDATE tariffs SET unlock_price = $1, minute_price = $2 WHERE id = $3`
	cmdTag, err := pool.Exec(context.Background(), query, unlockPrice, minutePrice, id)
	if err != nil {
		return fmt.Errorf("помилка оновлення тарифу: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("тариф не знайдено")
	}
	return nil
}

// AddTariff додає новий тариф
func AddTariff(pool *pgxpool.Pool, name, vehicleType string, unlockPrice, minutePrice float64) (int, error) {
	query := `INSERT INTO tariffs (name, type, unlock_price, minute_price) VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	if err := pool.QueryRow(context.Background(), query, name, vehicleType, unlockPrice, minutePrice).Scan(&id); err != nil {
		return 0, fmt.Errorf("помилка створення тарифу: %w", err)
	}
	return id, nil
}

// DeleteTariff видаляє тариф з бази
func DeleteTariff(pool *pgxpool.Pool, id int) error {
	query := `DELETE FROM tariffs WHERE id = $1`
	cmdTag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("помилка видалення (можливо, цей тариф зараз використовується транспортом): %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("тариф не знайдено")
	}
	return nil
}

// GetAllUsersAdmin повертає список усіх користувачів
func GetAllUsersAdmin(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `SELECT id, name, email, phone, role, balance, is_blocked FROM users WHERE deleted_at IS NULL ORDER BY id DESC`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту користувачів: %w", err)
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var name, email, phone, role string
		var balance float64
		var isBlocked bool

		if err := rows.Scan(&id, &name, &email, &phone, &role, &balance, &isBlocked); err != nil {
			return nil, err
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"name":       name,
			"email":      email,
			"phone":      phone,
			"role":       role,
			"balance":    balance,
			"is_blocked": isBlocked,
		})
	}
	return users, nil
}

// ToggleUserBlock блокує або розблоковує користувача (не дозволяє блокувати адмінів)
func ToggleUserBlock(pool *pgxpool.Pool, userID int, isBlocked bool) error {
	query := `UPDATE users SET is_blocked = $1 WHERE id = $2 AND role != 'admin' AND deleted_at IS NULL`
	if _, err := pool.Exec(context.Background(), query, isBlocked, userID); err != nil {
		return fmt.Errorf("помилка оновлення статусу користувача: %w", err)
	}
	return nil
}
