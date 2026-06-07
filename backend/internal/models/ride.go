package models

import "time"

// Ride описує сутність поїздки в системі
type Ride struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	VehicleID  int        `json:"vehicle_id"`
	Status     string     `json:"status"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"` // Вказівник, бо на старті тут буде NULL
	TotalPrice float64    `json:"total_price"`
}

// StartRideRequest описує ID самоката, який юзер хоче орендувати
type StartRideRequest struct {
	VehicleID int `json:"vehicle_id"`
}
