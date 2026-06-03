package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-rental/internal/database"

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
		// 1. Витягуємо ID з URL (наприклад, з /api/vehicles/1 витягне "1")
		idStr := chi.URLParam(r, "id")

		// 2. Перетворюємо рядок у число
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Некоректний ID самоката", http.StatusBadRequest)
			return
		}

		// 3. Йдемо в базу за конкретним самокатом
		vehicle, err := database.GetVehicleByID(dbPool, id)
		if err != nil {
			// Якщо база повернула помилку, скоріше за все самокат не знайдено
			http.Error(w, "Самокат не знайдено", http.StatusNotFound)
			return
		}

		// 4. Віддаємо успішну відповідь у форматі JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicle)
	}
}
