package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port   string
	DBConn string
}

func LoadConfig() *Config {
	port := getEnv("PORT", "3000")

	// Configuraciones locales por defecto para Docker Compose
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "novobanco")

	// Formato estándar para lib/pq
	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	return &Config{
		Port:   port,
		DBConn: dbConn,
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
