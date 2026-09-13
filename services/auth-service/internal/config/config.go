package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds runtime settings for auth-service.
type Config struct {
	HTTPPort       string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	RabbitURL      string
}

// Load reads config from environment variables with simple defaults for local run.
func Load() Config {
	expireHours, err := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	if err != nil {
		expireHours = 24
	}

	return Config{
		HTTPPort:       getEnv("HTTP_PORT", "8081"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://auth:auth@localhost:5432/auth_db?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpireHours: expireHours,
		RabbitURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

// TokenTTL returns JWT lifetime.
func (c Config) TokenTTL() time.Duration {
	return time.Duration(c.JWTExpireHours) * time.Hour
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
