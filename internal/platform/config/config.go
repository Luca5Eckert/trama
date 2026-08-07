package config

import (
	"log/slog"
	"os"
)

type Config struct {
	HTTPAddress string
	LogLevel    slog.Level
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddress: valueOrDefault("HTTP_ADDRESS", ":8080"),
		LogLevel:    logLevel(valueOrDefault("LOG_LEVEL", "info")),
		DatabaseURL: valueOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/trama?sslmode=disable"),
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func logLevel(value string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return slog.LevelInfo
	}
	return level
}
