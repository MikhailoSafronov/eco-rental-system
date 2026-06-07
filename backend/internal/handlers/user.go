package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/auth"
	"eco-rental/internal/database"
	"eco-rental/internal/middleware"
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

// GetProfile повертає дані профілю поточного авторизованого користувача
func GetProfile(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Дістаємо ID користувача з контексту запиту
		userIDRaw := r.Context().Value(middleware.UserIDKey)

		// Перевіряємо, чи дійсно це число (int)
		userID, ok := userIDRaw.(int)
		if !ok {
			http.Error(w, "Помилка авторизації: неможливо прочитати ID", http.StatusUnauthorized)
			return
		}

		// 2. Йдемо в базу даних за повною інформацією
		user, err := database.GetUserByID(pool, userID)
		if err != nil {
			http.Error(w, "Користувача не знайдено", http.StatusNotFound)
			return
		}

		// 3. Віддаємо дані клієнту у форматі JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

// TopUpBalance обробляє запит на поповнення віртуального рахунку (НОВЕ)
func TopUpBalance(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Витягуємо ID користувача з контексту так само, як у GetProfile
		userIDRaw := r.Context().Value(middleware.UserIDKey)

		userID, ok := userIDRaw.(int)
		if !ok {
			http.Error(w, "Помилка авторизації: неможливо прочитати ID", http.StatusUnauthorized)
			return
		}

		// Декодуємо JSON із сумою
		var req models.TopUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некоректний формат даних", http.StatusBadRequest)
			return
		}

		// Перевіряємо, щоб сума була адекватною
		if req.Amount <= 0 {
			http.Error(w, "Сума поповнення має бути більшою за нуль", http.StatusBadRequest)
			return
		}

		// Оновлюємо баланс у базі даних
		err := database.AddUserBalance(dbPool, userID, req.Amount)
		if err != nil {
			http.Error(w, "Не вдалося поповнити баланс", http.StatusInternalServerError)
			return
		}

		// Відповідаємо успіхом
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success", "message":"Баланс успішно поповнено"}`))
	}
}
