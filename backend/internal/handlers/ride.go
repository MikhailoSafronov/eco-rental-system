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

// EndRide обробляє запит на завершення поїздки
func EndRide(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Витягуємо ID користувача з токена так само, як і під час старту
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			http.Error(w, "Помилка авторизації: неможливо прочитати ID", http.StatusUnauthorized)
			return
		}

		// 2. Читаємо JSON з URL фотографії паркування (ОБОВ'ЯЗКОВО)
		var req struct {
			EndPhotoURL string `json:"end_photo_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Помилка читання JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.EndPhotoURL == "" {
			http.Error(w, "Для завершення поїздки обов'язково потрібно додати фотографію паркування (поле end_photo_url)", http.StatusBadRequest)
			return
		}

		// 3. Викликаємо базу даних і передаємо URL фото
		ride, err := database.EndRide(pool, userID, req.EndPhotoURL)
		if err != nil {
			// Якщо поїздки немає або сталась помилка при розрахунках
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 3. Повертаємо успішну відповідь із фінальним чеком
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Поїздку завершено! Дякуємо, що обрали нас 💚",
			"receipt": ride, // Повертаємо чек із сумою та часом
		})
	}
}

// GetRideHistory повертає історію поїздок поточного користувача
func GetRideHistory(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDRaw := r.Context().Value(middleware.UserIDKey)
		userID, ok := userIDRaw.(int)
		if !ok {
			http.Error(w, "Помилка авторизації: неможливо прочитати ID", http.StatusUnauthorized)
			return
		}

		history, err := database.GetUserRideHistory(pool, userID)
		if err != nil {
			http.Error(w, "Помилка отримання історії поїздок", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rides": history,
		})
	}
}
