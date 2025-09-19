package config

import (
	"log"
	"os"
	"strconv"

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
		AppPort:     getEnv("APP_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "your-default-jwt-secret-change-in-production"), // Add this
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

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Invalid integer value for %s: %s. Using default: %d", key, value, defaultValue)
		return defaultValue
	}

	return intValue
}
