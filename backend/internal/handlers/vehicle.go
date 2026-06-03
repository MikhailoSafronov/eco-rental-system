package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-rental/internal/database"
	"eco-rental/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAvailableVehicles повертає клієнту список вільних самокатів
func GetAvailableVehicles(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vehicles, err := database.GetAllAvailableVehicles(dbPool)
		if err != nil {
			http.Error(w, "Помилка отримання даних з БД", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicles)
	}
}

// GetVehicle повертає інформацію про конкретний самокат за його ID
func GetVehicle(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Некоректний ID самоката", http.StatusBadRequest)
			return
		}

		vehicle, err := database.GetVehicleByID(dbPool, id)
		if err != nil {
			http.Error(w, "Самокат не знайдено", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicle)
	}
}

// UpdateTelemetry приймає дані від фізичного самоката (IoT)
func UpdateTelemetry(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Витягуємо UUID з URL
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			http.Error(w, "UUID обов'язковий", http.StatusBadRequest)
			return
		}

		// Декодуємо JSON від самоката
		var req models.TelemetryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некоректний формат даних", http.StatusBadRequest)
			return
		}

		// Базова валідація батареї
		if req.BatteryLevel < 0 || req.BatteryLevel > 100 {
			http.Error(w, "Некоректний рівень заряду", http.StatusBadRequest)
			return
		}

		// Оновлюємо базу даних
		err := database.UpdateVehicleTelemetry(dbPool, uuid, req.Latitude, req.Longitude, req.BatteryLevel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Відповідаємо самокату, що все успішно (200 OK)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"telemetry updated"}`))
	}
}
