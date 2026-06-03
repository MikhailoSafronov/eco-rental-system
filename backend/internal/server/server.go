package server

import (
	"net/http"

	"eco-rental/internal/handlers"
	"eco-rental/internal/middleware"

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

	// Маршрути для транспорту
	r.Get("/api/vehicles", handlers.GetAvailableVehicles(dbPool))
	r.Get("/api/vehicles/{id}", handlers.GetVehicle(dbPool)) // НОВИЙ МАРШРУТ ДЛЯ ОДНОГО САМОКАТА

	// 2. Захищені маршрути (тільки з токеном)
	r.Group(func(r chi.Router) {
		// Чіпляємо нашого "охоронця" на цю групу
		r.Use(middleware.Auth)

		// Реальний маршрут отримання профілю
		r.Get("/api/users/me", handlers.GetProfile(dbPool))
	})

	return r
}
