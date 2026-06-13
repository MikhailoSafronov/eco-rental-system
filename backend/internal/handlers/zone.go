package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetParkingZones повертає клієнту список активних паркувальних зон
func GetParkingZones(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zones, err := database.GetActiveParkingZones(dbPool)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка отримання паркувальних зон")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(zones)
	}
}
