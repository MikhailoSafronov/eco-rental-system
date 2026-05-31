package main

import (
	"log"
	"net/http" // Вбудований пакет Go для роботи з HTTP

	"eco-rental/internal/database"
	"eco-rental/internal/server"
)

func main() {
	// 1. Підключаємось до бази даних PostgreSQL
	dsn := "postgres://admin:super_password@localhost:5433/eco_rental"
	dbPool, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("❌ Критична помилка БД: %v\n", err)
	}
	defer dbPool.Close()

	// 2. Ініціалізуємо наш роутер і передаємо туди пул бази даних (він нам знадобиться пізніше)
	router := server.RegisterRoutes(dbPool)

	// 3. Запускаємо веб-сервер на порту 8080
	log.Println("🌐 Веб-сервер запущено на порту :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("❌ Помилка запуску сервера: %v", err)
	}
}
