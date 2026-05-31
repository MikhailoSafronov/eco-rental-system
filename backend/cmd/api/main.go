package main

import (
	"log"

	"eco-rental/internal/database"
)

func main() {
	// DSN (Data Source Name)
	// Формат: postgres://користувач:пароль@хост:порт/назва_бд
	dsn := "postgres://admin:super_password@localhost:5433/eco_rental"

	// Викликаємо функцію підключення
	dbPool, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("❌ Критична помилка: %v\n", err)
	}
	// Гарантуємо закриття з'єднання при вимкненні програми
	defer dbPool.Close()

	log.Println("🚀 Сервер еко-оренди готовий до роботи!")
}
