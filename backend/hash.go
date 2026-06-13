package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Генеруємо правильний криптографічний хеш для пароля password123
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	fmt.Println(string(hash))
}
