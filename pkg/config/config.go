package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     string
	JWTSecret   string
	RedisURL    string
	Environment string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment values")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"),
		AppPort:     getEnv("AUTH_SERVICE_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "Hello_world"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv returns the value of the environment variable or fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
