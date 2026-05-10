package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	filePath = "backend/internal/database/seeds/genres.json"
)

func normalizeGenreName(name string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))

	formattedTag := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, lowerName)

	return formattedTag
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load("backend/.env"); err != nil {
			logger.Warn("could not load .env file, relying on environment variables", "error", err)
		}
	}
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		logger.Error("DATABASE_URL is not set in .env or environment variables")
		os.Exit(1)
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("Successfully connected to Supabase")

	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read genres.json", "error", err)
	}

	var genresNames []string
	if err := json.Unmarshal(data, &genresNames); err != nil {
		logger.Error("Failed to unmarshal genres", "error", err)
	}

	batch := &pgx.Batch{}

	for _, genreName := range genresNames {
		normalized := normalizeGenreName(genreName)
		if normalized == "" {
			logger.Warn("Skipping genre with empty normalized name", "original", genreName)
			continue
		}
		batch.Queue("INSERT INTO genres (name, normalized_name) VALUES ($1, $2) ON CONFLICT (normalized_name) DO NOTHING", genreName, normalized)

	}

	logger.Info("Inserting genres into the database...")
	br := dbPool.SendBatch(ctx, batch)
	defer br.Close()

	insertedCount := 0
	for i := 0; i < batch.Len(); i++ {
		tag, err := br.Exec()
		if err != nil {
			logger.Error("Failed to execute batch insert for genre", "error", err)
		}
		insertedCount += int(tag.RowsAffected())
	}

	logger.Info("Seeding complete! Successfully inserted new genres.", "count", insertedCount)
}
