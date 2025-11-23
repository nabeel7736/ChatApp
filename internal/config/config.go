package config

import (
	"os"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSL     string
	JWTSecret string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASS", "sanufinu786"),
		DBName:    getEnv("DB_NAME", "chatapp"),
		DBSSL:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret: getEnv("JWT_SECRET", "secret123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
