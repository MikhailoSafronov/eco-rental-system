package models

// Vehicle представляє транспортний засіб, який ми віддаємо клієнту
type Vehicle struct {
	ID           int     `json:"id"`
	UUID         string  `json:"uuid"`
	BatteryLevel int     `json:"battery_level"`
	Status       string  `json:"status"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}
