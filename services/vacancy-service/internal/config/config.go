package config

import (
	"os"
)

// Config holds runtime settings for vacancy-service.
type Config struct {
	HTTPPort           string
	DatabaseURL        string
	JWTSecret          string
	RabbitURL          string
	EmployerServiceURL string
}

// Load reads config from environment variables.
func Load() Config {
	return Config{
		HTTPPort:           getEnv("HTTP_PORT", "8084"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://vacancy:vacancy@localhost:5432/vacancy_db?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		RabbitURL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		EmployerServiceURL: getEnv("EMPLOYER_SERVICE_URL", "http://localhost:8083"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
