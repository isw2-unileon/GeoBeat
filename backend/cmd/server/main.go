// Package main is the entry point for the backend server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/isw2-unileon/GeoBeat/backend/internal/config"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/isw2-unileon/GeoBeat/backend/internal/oauth"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/tools"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	// TODO: launch chron job to generate daily challenge at midnight UTC

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL is not set in .env or environment variables")
		os.Exit(1)
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("Successfully connected to Supabase")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			slog.Error("Failed to write health check response", "error", err)
		}
	})

	// ADD ENDPOINTS

	userRepo := pgdb.NewPostgresUserRepo(dbPool)
	tokenizer := tools.NewJWTTokenizer("sercreto secretisimo")
	hasher := tools.NewBCryptHasher()
	google_provider := oauth.NewGoogleOAuthProvider(cfg.GoogleClientID, cfg.GoogleSecret, cfg.RedirectURL+string(geouser.ProviderGoogle))
	service_providers := map[geouser.AuthProvider]service.OAuthProvider{
		geouser.ProviderGoogle: google_provider,
	}
	server_providers := map[geouser.AuthProvider]server.OAuthProvider{
		geouser.ProviderGoogle: google_provider,
	}
	authService := service.NewAuthService(userRepo, tokenizer, hasher, service_providers)
	authHandler := server.NewAuthHandler(authService, server_providers, cfg)

	authHandler.RegisterRoutes(mux)

	corsHandler := server.CorsMiddleware(cfg, mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server crashed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server stopped cleanly")
}
