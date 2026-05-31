package models

import (
	"time"
)

// User - це точне відображення рядка з нашої таблиці users у базі даних
type User struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	PasswordHash string     `json:"-"` // ТИРЕ ОЗНАЧАЄ: ніколи не віддавати це поле в JSON (безпека!)
	Role         string     `json:"role"`
	Balance      float64    `json:"balance"`
	IsBlocked    bool       `json:"is_blocked"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"` // Вказівник (*), бо в БД це поле може бути NULL
}

// RegisterUserRequest - це структура даних, яку ми очікуємо отримати з Postman
type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"` // Відкритий пароль від користувача (який ми захешуємо)
}
