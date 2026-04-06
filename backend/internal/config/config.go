package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port            string
	GinMode         string
	CORSAllowOrigin string
	LastFMAPIKey    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	godotenv.Load("backend/.env") // Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's fine if the .env file doesn't exist, we'll just rely on actual environment variables
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
		LastFMAPIKey:    getEnv("LASTFM_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
