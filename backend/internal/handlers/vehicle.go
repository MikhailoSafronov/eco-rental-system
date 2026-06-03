package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetAvailableVehicles повертає клієнту список вільних самокатів
func GetAvailableVehicles(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Йдемо в базу даних за самокатами
		vehicles, err := database.GetAllAvailableVehicles(dbPool)
		if err != nil {
			http.Error(w, "Помилка отримання даних з БД", http.StatusInternalServerError)
			return
		}

		// Віддаємо успішну відповідь у форматі JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicles)
	}
}
