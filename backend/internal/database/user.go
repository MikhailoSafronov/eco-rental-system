package database

import (
	"context"
	"eco-rental/internal/models" // Імпортуємо твої щойно створені моделі

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser додає нового користувача в базу даних
func CreateUser(pool *pgxpool.Pool, req models.RegisterUserRequest, hashedPassword string) (int64, error) {
	// SQL-запит для вставки. Зверни увагу на $1, $2... — це захист від SQL-ін'єкцій
	// RETURNING id — це геніальна фішка PostgreSQL, яка одразу повертає ID створеного рядка
	query := `
		INSERT INTO users (name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	var newID int64

	// Виконуємо запит. QueryRow — бо очікуємо повернути рівно один рядок (з нашим новим ID)
	err := pool.QueryRow(
		context.Background(),
		query,
		req.Name,
		req.Email,
		req.Phone,
		hashedPassword,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}
