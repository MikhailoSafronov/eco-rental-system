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

// GetAvailableVehicles повертає клієнту список вільного транспорту
func GetAvailableVehicles(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vehicles, err := database.GetAllAvailableVehicles(dbPool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка отримання даних з БД")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicles)
	}
}

// GetVehicle повертає інформацію про конкретний транспортний засіб за його ID
func GetVehicle(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний ID транспорту")
			return
		}

		vehicle, err := database.GetVehicleByID(dbPool, id)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, "Транспорт не знайдено")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicle)
	}
}

// UpdateTelemetry приймає дані від фізичного транспорту (IoT)
func UpdateTelemetry(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Витягуємо UUID з URL
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			RespondWithError(w, http.StatusBadRequest, "UUID обов'язковий")
			return
		}

		// Декодуємо JSON від транспорту
		var req models.TelemetryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний формат даних")
			return
		}

		// Базова валідація батареї
		if req.BatteryLevel < 0 || req.BatteryLevel > 100 {
			RespondWithError(w, http.StatusBadRequest, "Некоректний рівень заряду")
			return
		}

		// Оновлюємо базу даних
		err := database.UpdateVehicleTelemetry(dbPool, uuid, req.Latitude, req.Longitude, req.BatteryLevel)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Відповідаємо транспорту, що все успішно (200 OK)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","message":"telemetry updated"}`))
	}
}
