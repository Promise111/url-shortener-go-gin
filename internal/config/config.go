package config

import (
	"log/slog"
	"os"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() (*Config, error) {
	var err error
	err = dotenv.Load()
	if err != nil {
		slog.Error("Warning: ", "message", ".env file not found "+err.Error())
		return nil, err
	}

	var config *Config = &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
	}

	return config, err
}
