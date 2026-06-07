package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/database"
	"eco-rental/internal/middleware"
	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartRide обробляє HTTP-запит на початок оренди
func StartRide(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Дістаємо ID користувача з токена через контекст
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			http.Error(w, "Помилка авторизації: неможливо прочитати ID", http.StatusUnauthorized)
			return
		}

		// 2. Декодуємо вхідний JSON
		var req models.StartRideRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Невалідний JSON запиту", http.StatusBadRequest)
			return
		}

		// 3. Запускаємо бізнес-логіку оренди
		ride, err := database.StartRide(pool, userID, req.VehicleID)
		if err != nil {
			// Якщо бізнес-логіка повернула помилку (мало грошей, самокат зайнятий), віддаємо її клієнту
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 4. Повертаємо успішну відповідь
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Поїздку успішно розпочато! Приємної подорожі 🛴",
			"ride":    ride,
		})
	}
}
