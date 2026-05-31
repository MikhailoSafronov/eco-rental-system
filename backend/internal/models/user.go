package models

// RegisterUserRequest описує тіло запиту для реєстрації
type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// LoginUserRequest описує тіло запиту для входу
type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User описує повну модель користувача в системі
type User struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	PasswordHash string  `json:"-"` // Гарантує, що пароль не потрапить у JSON-відповідь
	Role         string  `json:"role"`
	Balance      float64 `json:"balance"`
	IsBlocked    bool    `json:"is_blocked"`
}
