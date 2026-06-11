package database

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

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

// RunMigrations автоматично перевіряє та накатує нові міграції при старті сервера
func RunMigrations(dsn string) error {
	log.Println("🔄 Перевірка міграцій бази даних...")

	m, err := migrate.New(
		"file://migrations", // Шлях до папки migrations
		dsn,
	)
	if err != nil {
		return fmt.Errorf("помилка ініціалізації міграцій: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("помилка застосування міграцій: %w", err)
	}

	log.Println("✅ Міграції бази даних в актуальному стані!")
	return nil
}
