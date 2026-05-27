package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	MigrateURL  string
	JWTSecret   string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(key + " is required")
	}
	return value
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8082"),
		DatabaseURL: getEnv("POSTGRES_URL", "postgres://postgres:secret@localhost:5432/mixfood_menu?sslmode=disable"),
		MigrateURL:  getEnv("MIGRATE_URL", "file://migrations"),
		JWTSecret:   getEnvRequired("JWT_SECRET"),
	}
}
