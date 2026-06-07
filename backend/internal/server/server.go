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

	// Маршрути для транспорту (Клієнтські)
	r.Get("/api/vehicles", handlers.GetAvailableVehicles(dbPool))
	r.Get("/api/vehicles/{id}", handlers.GetVehicle(dbPool))

	// IoT ендпоінт для заліза (Прихований)
	r.Patch("/api/iot/vehicles/{uuid}/telemetry", handlers.UpdateTelemetry(dbPool))

	// 2. Захищені маршрути (тільки з токеном)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		// Користувач та фінанси
		r.Get("/api/users/me", handlers.GetProfile(dbPool))
		r.Post("/api/users/topup", handlers.TopUpBalance(dbPool))

		// Поїздки (НОВИЙ МАРШРУТ)
		r.Post("/api/rides/start", handlers.StartRide(dbPool))
	})

	return r
}
