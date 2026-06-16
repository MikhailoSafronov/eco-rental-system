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
	r.Get("/api/zones", handlers.GetParkingZones(dbPool))

	// IoT ендпоінт для заліза (Прихований)
	r.Patch("/api/iot/vehicles/{uuid}/telemetry", handlers.UpdateTelemetry(dbPool))

	// Роздача статичних файлів з папки uploads (фотографії паркування)
	fs := http.FileServer(http.Dir("uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))

	// 2. Захищені маршрути (тільки з токеном)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		// Користувач та фінанси
		r.Get("/api/users/me", handlers.GetProfile(dbPool))
		r.Post("/api/users/topup", handlers.TopUpBalance(dbPool))
		r.Get("/api/users/payments", handlers.GetPaymentHistory(dbPool))

		// Поїздки
		r.Post("/api/rides/start", handlers.StartRide(dbPool))
		r.Post("/api/rides/end", handlers.EndRide(dbPool)) // <-- ДОДАНИЙ РЯДОК ЗАВЕРШЕННЯ ПОЇЗДКИ
		r.Get("/api/rides/history", handlers.GetRideHistory(dbPool))

		// Завантаження файлів
		r.Post("/api/upload", handlers.UploadPhoto())
	})

	// 3. Адміністративні маршрути
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)                 // Спочатку перевіряємо наявність токена
		r.Use(middleware.RequireAdmin(dbPool)) // Потім перевіряємо роль адміністратора

		r.Get("/api/admin/vehicles", handlers.GetAllVehiclesAdmin(dbPool))
		r.Post("/api/admin/vehicles", handlers.AddVehicle(dbPool))
		r.Patch("/api/admin/vehicles/{id}/status", handlers.UpdateVehicleStatus(dbPool))
		r.Get("/api/admin/models", handlers.GetAllModelsAdmin(dbPool))
		r.Get("/api/admin/tariffs", handlers.GetAllTariffsAdmin(dbPool))
		r.Post("/api/admin/zones", handlers.AddParkingZone(dbPool))
		r.Delete("/api/admin/vehicles/{id}", handlers.DeleteVehicle(dbPool))
		r.Delete("/api/admin/zones/{id}", handlers.DeleteParkingZone(dbPool))
		r.Patch("/api/admin/tariffs/{id}", handlers.UpdateTariff(dbPool))
		r.Post("/api/admin/tariffs", handlers.AddTariff(dbPool))
		r.Delete("/api/admin/tariffs/{id}", handlers.DeleteTariff(dbPool))
		r.Get("/api/admin/users", handlers.GetAllUsersAdmin(dbPool))
		r.Patch("/api/admin/users/{id}/block", handlers.ToggleUserBlock(dbPool))
		r.Get("/api/admin/rides", handlers.GetAllRidesAdmin(dbPool))
		r.Post("/api/admin/models", handlers.AddVehicleModel(dbPool))
		r.Delete("/api/admin/models/{id}", handlers.DeleteVehicleModel(dbPool))
	})

	return r
}
