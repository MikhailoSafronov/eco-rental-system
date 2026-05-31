package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect створює пул з'єднань з базою даних і повертає його
func Connect(connString string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("помилка парсингу конфігурації: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("помилка підключення до БД: %w", err)
	}

	// Перевіряємо, чи база реально відповідає
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("база не відповідає на пінг: %w", err)
	}

	log.Println("✅ Успішно підключено до бази даних PostgreSQL!")
	return pool, nil
}
