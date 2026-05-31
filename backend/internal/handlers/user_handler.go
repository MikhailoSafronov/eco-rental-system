package handlers

import (
	"encoding/json"
	"net/http"

	"eco-rental/internal/database"
	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser створює обробник для реєстрації нових користувачів
func RegisterUser(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Читаємо JSON з Postman і перекладаємо його в наш "Бланк" (RegisterUserRequest)
		var req models.RegisterUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Неправильний формат даних", http.StatusBadRequest)
			return
		}

		// 2. Валідація: перевіряємо, щоб юзер не надіслав порожні поля
		if req.Email == "" || req.Password == "" || req.Name == "" || req.Phone == "" {
			http.Error(w, "Всі поля є обов'язковими", http.StatusBadRequest)
			return
		}

		// 3. Шифруємо пароль алгоритмом bcrypt (10 - це стандартний рівень складності хешування)
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			http.Error(w, "Помилка шифрування пароля", http.StatusInternalServerError)
			return
		}
		hashedPassword := string(hashedBytes)

		// 4. Передаємо дані "Кухарю" — зберігаємо в PostgreSQL
		newID, err := database.CreateUser(pool, req, hashedPassword)
		if err != nil {
			// Якщо email або телефон вже існують, база видасть помилку (бо в нас стоїть UNIQUE)
			http.Error(w, "Помилка збереження в базу. Можливо, такий email вже існує", http.StatusInternalServerError)
			return
		}

		// 5. Формуємо успішну відповідь для клієнта (Postman)
		w.Header().Set("Content-Type", "application/json") // Кажемо, що віддаємо JSON
		w.WriteHeader(http.StatusCreated)                  // Ставимо статус 201 Created

		// Відправляємо гарний JSON з ID нового користувача
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Користувача успішно створено 🚀",
			"user_id": newID,
		})
	}
}
