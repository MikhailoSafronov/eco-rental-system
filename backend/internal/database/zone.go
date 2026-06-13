package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetActiveParkingZones повертає паркувальні зони у форматі GeoJSON
func GetActiveParkingZones(pool *pgxpool.Pool) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			id, name, ST_AsGeoJSON(polygon)::json AS geojson
		FROM parking_zones
		WHERE is_active = true
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("помилка запиту до БД: %w", err)
	}
	defer rows.Close()

	zones := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var name string
		var geojsonBytes []byte

		if err := rows.Scan(&id, &name, &geojsonBytes); err != nil {
			return nil, fmt.Errorf("помилка сканування рядка: %w", err)
		}

		var geojson map[string]interface{}
		if err := json.Unmarshal(geojsonBytes, &geojson); err != nil {
			return nil, fmt.Errorf("помилка парсингу GeoJSON: %w", err)
		}

		zones = append(zones, map[string]interface{}{"id": id, "name": name, "geojson": geojson})
	}

	return zones, nil
}
