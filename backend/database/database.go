package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var DB *sql.DB

// InitDB инициализирует подключение к базе данных
func InitDB() error {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения системы")
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "fitness_club")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка ping БД: %v", err)
	}

	log.Println("Успешное подключение к базе данных PostgreSQL")
	return nil
}

// CloseDB закрывает подключение к базе данных
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Подключение к базе данных закрыто")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

