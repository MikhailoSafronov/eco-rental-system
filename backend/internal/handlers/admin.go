package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
			Battery   int     `json:"battery_level"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON запиту")
			return
		}

		id, uuid, err := database.AddVehicle(pool, req.ModelID, req.TariffID, req.Battery, req.Latitude, req.Longitude)
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

// GetAllModelsAdmin повертає список моделей для створення транспорту
func GetAllModelsAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := database.GetAllModelsAdmin(pool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models)
	}
}

// GetAllTariffsAdmin повертає список тарифів для створення транспорту
func GetAllTariffsAdmin(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tariffs, err := database.GetAllTariffsAdmin(pool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tariffs)
	}
}

// AddParkingZone обробляє створення нової паркувальної зони
func AddParkingZone(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name   string      `json:"name"`
			Points [][]float64 `json:"points"` // Очікується масив пар: [[lon, lat], [lon, lat], ...]
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON")
			return
		}

		if len(req.Points) < 3 {
			RespondWithError(w, http.StatusBadRequest, "Полігон зони повинен мати мінімум 3 точки")
			return
		}

		// PostGIS вимагає, щоб полігон був "замкненим" (перша і остання точка мають співпадати)
		firstPt := req.Points[0]
		lastPt := req.Points[len(req.Points)-1]
		if firstPt[0] != lastPt[0] || firstPt[1] != lastPt[1] {
			req.Points = append(req.Points, firstPt)
		}

		// Формуємо рядок WKT (Well-Known Text): "POLYGON((lon1 lat1, lon2 lat2, ...))"
		var pointStrs []string
		for _, p := range req.Points {
			pointStrs = append(pointStrs, fmt.Sprintf("%f %f", p[0], p[1]))
		}
		wktPolygon := fmt.Sprintf("POLYGON((%s))", strings.Join(pointStrs, ", "))

		id, err := database.AddParkingZone(pool, req.Name, wktPolygon)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Зону успішно додано",
			"id":      id,
		})
	}
}
