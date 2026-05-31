package server

import (
	"net/http"

	"eco-rental/internal/handlers" // ДОДАЛИ: імпорт наших хендлерів

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter створює і налаштовує наш HTTP-роутер
func SetupRouter(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Старий тестовий маршрут
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK! Сервер живий і готовий приймати запити 🚀"))
	})

	// НОВИЙ БЛОК: Маршрути для користувачів
	// Ми групуємо всі запити, які починаються з /api/users
	r.Route("/api/users", func(r chi.Router) {
		// Кажемо роутеру: якщо прийшов POST запит на /register,
		// передай керування нашому "Офіціанту" і дай йому доступ до бази (db)
		r.Post("/register", handlers.RegisterUser(db))
	})

	return r
}
