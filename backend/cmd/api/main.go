package main

import (
	"log"
	"net/http" // Вбудований пакет Go для роботи з HTTP
	"os"       // Пакет для роботи зі змінними середовища

	"eco-rental/internal/database"
	"eco-rental/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Завантажуємо змінні з файлу .env, який лежить на рівень вище (у корені проекту)
	// Якщо файлу немає (наприклад, на продакшен-сервері), код просто піде далі
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("⚠️ Файл .env не знайдено, використовуються системні змінні")
	}

	// 2. Читаємо секрети та перевіряємо їх наявність
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("❌ Критична помилка: змінна DB_DSN не встановлена")
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("❌ Критична помилка: змінна JWT_SECRET не встановлена")
	}

	// 3. Підключаємось до бази даних PostgreSQL, використовуючи рядок з .env
	dbPool, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("❌ Критична помилка БД: %v\n", err)
	}
	defer dbPool.Close() // Гарантуємо закриття з'єднання при вимкненні сервера

	// 4. Ініціалізуємо наш роутер і передаємо туди пул бази даних
	router := server.RegisterRoutes(dbPool)

	// 5. Запускаємо веб-сервер на порту 8080
	log.Println("🌐 Веб-сервер запущено на порту :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("❌ Помилка запуску сервера: %v", err)
	}
}
