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
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка шифрування пароля")
			return
		}

		newID, err := database.CreateUser(pool, req, string(hashedPassword))
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка збереження в базу. Можливо, такий email вже існує")
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
			RespondWithError(w, http.StatusBadRequest, "Невалідний JSON")
			return
		}

		// Шукаємо користувача в БД
		user, err := database.GetUserByEmail(pool, req.Email)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Невірний email або пароль")
			return
		}

		// Порівнюємо хеш із БД та введений пароль
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Невірний email або пароль")
			return
		}

		// Генеруємо JWT токен
		tokenString, err := auth.GenerateToken(user.ID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Помилка генерації токена")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Успішний вхід",
			"user_id": user.ID,
			"token":   tokenString, // Віддаємо токен клієнту
			"role":    user.Role,   // Віддаємо роль для перевірки доступу на фронтенді
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
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації: неможливо прочитати ID")
			return
		}

		// 2. Йдемо в базу даних за повною інформацією
		user, err := database.GetUserByID(pool, userID)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, "Користувача не знайдено")
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
			RespondWithError(w, http.StatusUnauthorized, "Помилка авторизації: неможливо прочитати ID")
			return
		}

		// Декодуємо JSON із сумою
		var req models.TopUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Некоректний формат даних")
			return
		}

		// Перевіряємо, щоб сума була адекватною
		if req.Amount <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Сума поповнення має бути більшою за нуль")
			return
		}

		// Оновлюємо баланс у базі даних
		err := database.AddUserBalance(dbPool, userID, req.Amount)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Не вдалося поповнити баланс")
			return
		}

		// Відповідаємо успіхом
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Баланс успішно поповнено",
		})
	}
}
