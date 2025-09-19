package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment values")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"),
		AppPort:     getEnv("APP_PORT", "8080"),
	}
}

// getEnv returns the value of the environment variable or fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
