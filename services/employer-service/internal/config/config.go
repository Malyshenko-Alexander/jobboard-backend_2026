package config

import (
	"os"
)

// Config holds runtime settings for employer-service.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	JWTSecret   string
	RabbitURL   string
}

// Load reads config from environment variables.
func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "8083"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://employer:employer@localhost:5432/employer_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		RabbitURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
