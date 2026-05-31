package middleware

import (
	"context"
	"net/http"
	"strings"

	"eco-rental/internal/auth"
)

// Створюємо спеціальний тип для ключа контексту, щоб уникнути конфліктів
type contextKey string

const UserIDKey contextKey = "userID"

// Auth - це middleware для перевірки JWT токена
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Шукаємо заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Відсутній заголовок Authorization", http.StatusUnauthorized)
			return
		}

		// Токен має передаватися у форматі: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неправильний формат токена. Використовуйте 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Перевіряємо токен
		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Недійсний або прострочений токен", http.StatusUnauthorized)
			return
		}

		// Якщо все ок, записуємо userID у контекст запиту і пускаємо далі
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
