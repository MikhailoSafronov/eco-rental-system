package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eco-rental/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AddVehicle обробляє створення нового самоката адміністратором
func AddVehicle(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ModelID   int     `json:"model_id"`
			TariffID  int     `json:"tariff_id"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON запиту")
			return
		}

		id, uuid, err := database.AddVehicle(pool, req.ModelID, req.TariffID, req.Latitude, req.Longitude)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Самокат успішно додано",
			"vehicle": map[string]interface{}{"id": id, "uuid": uuid},
		})
	}
}

// UpdateVehicleStatus обробляє ручну зміну статусу (доступні: available, rented, low_battery, broken, maintenance)
func UpdateVehicleStatus(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний ID самоката")
			return
		}

		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
			RespondWithError(w, http.StatusBadRequest, "Необхідно вказати новий статус")
			return
		}

		if err := database.UpdateVehicleStatus(pool, id, req.Status); err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Статус самоката успішно оновлено"})
	}
}

// GetAllVehiclesAdmin повертає список усіх самокатів для дашборду адміністратора
func GetAllVehiclesAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vehicles, err := database.GetAllVehiclesAdmin(pool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicles)
	}
}
