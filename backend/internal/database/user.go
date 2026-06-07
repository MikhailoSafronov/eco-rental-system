package database

import (
	"context"
	"eco-rental/internal/models"
	"fmt" // Додано для форматування тексту помилок

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

// GetUserByID шукає користувача за його унікальним ідентифікатором
func GetUserByID(pool *pgxpool.Pool, id int) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, role, balance, is_blocked
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &models.User{}

	// Зверни увагу: ми не витягуємо password_hash, бо для профілю він не потрібен
	err := pool.QueryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.Balance,
		&user.IsBlocked,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// AddUserBalance додає вказану суму до існуючого балансу користувача (НОВЕ)
func AddUserBalance(pool *pgxpool.Pool, userID int, amount float64) error {
	query := `
		UPDATE users 
		SET balance = balance + $1 
		WHERE id = $2 AND deleted_at IS NULL
	`

	commandTag, err := pool.Exec(context.Background(), query, amount, userID)
	if err != nil {
		return fmt.Errorf("помилка поповнення балансу: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("користувача з ID %d не знайдено", userID)
	}

	return nil
}
