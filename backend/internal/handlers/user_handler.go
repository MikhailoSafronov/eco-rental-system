package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/auth"
	"eco-rental/internal/database"
	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Невалідний JSON", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Помилка шифрування пароля", http.StatusInternalServerError)
			return
		}

		newID, err := database.CreateUser(pool, req, string(hashedPassword))
		if err != nil {
			http.Error(w, "Помилка збереження в базу. Можливо, такий email вже існує", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Користувача успішно створено 🚀",
			"user_id": newID,
		})
	}
}

func LoginUser(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Невалідний JSON", http.StatusBadRequest)
			return
		}

		// Шукаємо користувача в БД
		user, err := database.GetUserByEmail(pool, req.Email)
		if err != nil {
			http.Error(w, "Невірний email або пароль", http.StatusUnauthorized)
			return
		}

		// Порівнюємо хеш із БД та введений пароль
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			http.Error(w, "Невірний email або пароль", http.StatusUnauthorized)
			return
		}

		// Генеруємо JWT токен
		tokenString, err := auth.GenerateToken(user.ID)
		if err != nil {
			http.Error(w, "Помилка генерації токена", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Успішний вхід",
			"user_id": user.ID,
			"token":   tokenString, // Віддаємо токен клієнту
		})
	}
}
