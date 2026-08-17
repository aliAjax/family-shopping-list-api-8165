package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	TokenTTL   time.Duration
}

func Load() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "mysql"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "shopping"),
		DBPassword: getEnv("DB_PASSWORD", "shopping_secret"),
		DBName:     getEnv("DB_NAME", "family_shopping"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
		TokenTTL:   time.Duration(getEnvInt("TOKEN_TTL_HOURS", 24)) * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
