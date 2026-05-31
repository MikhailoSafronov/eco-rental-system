package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Секретний ключ для підпису токенів.
// У реальному проекті він МАЄ зберігатися у змінних середовища (.env), а не в коді.
var jwtSecretKey = []byte("super_secret_key_for_eco_rental")

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
	return token.SignedString(jwtSecretKey)
}
