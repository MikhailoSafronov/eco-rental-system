package server

import (
	"fmt"
	"net/http"

	"eco-rental/internal/handlers"
	"eco-rental/internal/middleware" // Додаємо імпорт нашого охоронця

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// 1. Публічні маршрути (доступні всім)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK! Сервер живий і готовий приймати запити 🚀"))
	})

	r.Post("/api/users/register", handlers.RegisterUser(dbPool))
	r.Post("/api/users/login", handlers.LoginUser(dbPool))

	// 2. Захищені маршрути (тільки з токеном)
	r.Group(func(r chi.Router) {
		// Чіпляємо нашого "охоронця" на цю групу
		r.Use(middleware.Auth)

		// Тестовий захищений маршрут
		r.Get("/api/users/me", func(w http.ResponseWriter, r *http.Request) {
			// Дістаємо ID користувача, який охоронець поклав у контекст
			userID := r.Context().Value(middleware.UserIDKey)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"message": "Ти успішно пройшов охоронця!", "твоє_id": %v}`, userID)))
		})
	})

	return r
}
