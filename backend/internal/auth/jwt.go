package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getSecretKey динамічно дістає ключ зі змінних середовища
func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("Критична помилка: JWT_SECRET не встановлено")
	}
	return []byte(secret)
}

// GenerateToken створює JWT токен для користувача
func GenerateToken(userID int) (string, error) {
	// Створюємо "корисне навантаження" (payload) токена
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Токен дійсний 24 години
		"iat":     time.Now().Unix(),                     // Час створення
	}

	// Створюємо сам токен із вказанням алгоритму шифрування (HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Підписуємо токен нашим секретним ключем і перетворюємо у рядок
	return token.SignedString(getSecretKey())
}

// ValidateToken перевіряє токен і повертає user_id
func ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Перевіряємо, чи алгоритм шифрування той самий
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неочікуваний метод підпису")
		}
		return getSecretKey(), nil
	})

	if err != nil {
		return 0, err
	}

	// Дістаємо дані, якщо токен валідний
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// У JWT числа зберігаються як float64, тому робимо конвертацію
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("невалідний токен")
}
