package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/database"
	"eco-rental/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetPaymentHistory віддає клієнту його історію платежів
func GetPaymentHistory(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації")
			return
		}

		payments, err := database.GetUserPayments(dbPool, userID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"payments": payments})
	}
}
