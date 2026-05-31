package database

import (
	"context"
	"eco-rental/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser зберігає нового користувача в базу даних
func CreateUser(pool *pgxpool.Pool, req models.RegisterUserRequest, hashedPassword string) (int, error) {
	query := `
		INSERT INTO users (name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var newID int
	err := pool.QueryRow(context.Background(), query, req.Name, req.Email, req.Phone, hashedPassword).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// GetUserByEmail шукає користувача за email
func GetUserByEmail(pool *pgxpool.Pool, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role, balance, is_blocked
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &models.User{}

	err := pool.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Balance,
		&user.IsBlocked,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
