package auth

import (
	"errors"
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

// ValidateToken перевіряє токен і повертає user_id
func ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Перевіряємо, чи алгоритм шифрування той самий
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неочікуваний метод підпису")
		}
		return jwtSecretKey, nil
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
