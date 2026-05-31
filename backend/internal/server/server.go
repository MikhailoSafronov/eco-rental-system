package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter створює і налаштовує наш HTTP-роутер
func SetupRouter(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Мідлвари (проміжні обробники для зручності)
	r.Use(middleware.Logger)    // Красиво пише в консоль інформацію про кожен запит
	r.Use(middleware.Recoverer) // Захищає сервер від падіння, якщо десь станеться критична помилка (паніка)

	// Наш перший тестовий маршрут (ендпоінт)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK! Сервер живий і готовий приймати запити 🚀"))
	})

	return r
}
