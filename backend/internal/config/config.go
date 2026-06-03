package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port            string
	CORSAllowOrigin string
	LastFMAPIKey    string
	DatabaseURL     string
	GoogleClientID  string
	GoogleSecret    string
	RedirectURL     string
	FrontendURL     string
	JWTToken        string
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load("backend/.env"); err != nil {
			logger.Warn("could not load .env file, relying on environment variables", "error", err)
		}
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
		LastFMAPIKey:    getEnv("LASTFM_API_KEY", ""),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		GoogleClientID:  getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:    getEnv("GOOGLE_SECRET", ""),
		RedirectURL:     getEnv("REDIRECT_URL", ""),
		FrontendURL:     getEnv("FRONTEND_URL", ""),
		JWTToken:        mustGetEnv("JWT_TOKEN"), // TODO: Check if best practice
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic("Missing required environment variable: " + key)
}
