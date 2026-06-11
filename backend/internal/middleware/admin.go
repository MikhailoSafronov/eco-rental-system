package middleware

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RequireAdmin перевіряє, чи має користувач права адміністратора
func RequireAdmin(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Отримуємо ID користувача (встановлений попередньою мідлварою Auth)
			userIDRaw := r.Context().Value(UserIDKey)
			userID, ok := userIDRaw.(int)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Неавторизовано")
				return
			}

			// 2. Робимо запит до БД для отримання ролі
			var role string
			err := pool.QueryRow(r.Context(), "SELECT role FROM users WHERE id = $1 AND deleted_at IS NULL", userID).Scan(&role)
			if err != nil || role != "admin" {
				respondWithError(w, http.StatusForbidden, "Доступ заборонено: потрібні права адміністратора")
				return
			}

			// Якщо все добре, передаємо керування далі
			next.ServeHTTP(w, r)
		})
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}
