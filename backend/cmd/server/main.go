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
	"github.com/isw2-unileon/GeoBeat/backend/internal/lastfm"
	"github.com/isw2-unileon/GeoBeat/backend/internal/oauth"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/tools"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL is not set in .env or environment variables")
		os.Exit(1)
	}

	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to parse database URL", "error", err)
		os.Exit(1)
	}

	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("Successfully connected to Supabase")

	mux := setupRoutes(dbPool, cfg)

	corsHandler := server.CorsMiddleware(cfg, mux)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	startDailyChallengeJob(dbPool, cfg)

	runServer(srv, cfg.Port)
}

// setupRoutes wires all handlers and returns a fully configured ServeMux.
func setupRoutes(dbPool *pgxpool.Pool, cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	authHandler := buildAuthHandler(dbPool, cfg)
	dailyHandler := buildDailyHandler(dbPool)

	authHandler.RegisterRoutes(mux)
	dailyHandler.RegisterRoutes(mux, authHandler.AuthMiddleware)

	return mux
}

// buildAuthHandler constructs the auth handler and its dependencies.
func buildAuthHandler(dbPool *pgxpool.Pool, cfg *config.Config) *server.AuthHandler {
	userRepo := pgdb.NewPostgresUserRepo(dbPool)
	tokenRepo := pgdb.NewPostgresTokenRepo(dbPool)
	tokenizer := tools.NewJWTTokenizer(cfg.JWTToken)
	hasher := tools.NewBCryptHasher()

	googleProvider := oauth.NewGoogleOAuthProvider(
		cfg.GoogleClientID,
		cfg.GoogleSecret,
		cfg.RedirectURL+string(geouser.ProviderGoogle),
	)

	authService := service.NewAuthService(userRepo, tokenRepo, tokenizer, hasher,
		map[geouser.AuthProvider]service.OAuthProvider{
			geouser.ProviderGoogle: googleProvider,
		},
	)

	return server.NewAuthHandler(authService,
		map[geouser.AuthProvider]server.OAuthProvider{
			geouser.ProviderGoogle: googleProvider,
		},
		cfg,
	)
}

// buildDailyHandler constructs the daily challenge handler and its dependencies.
func buildDailyHandler(dbPool *pgxpool.Pool) *server.Handler {
	dailyRepo := pgdb.NewPostgresDailyRepo(dbPool)
	return server.NewHandler(service.NewService(dailyRepo, dailyRepo))
}

// startDailyChallengeJob schedules the daily challenge generation at midnight UTC.
func startDailyChallengeJob(dbPool *pgxpool.Pool, cfg *config.Config) {
	genreRepo := pgdb.NewPostgresGenreRepo(dbPool)
	musicProvider := lastfm.NewClient(cfg.LastFMAPIKey)
	dailyRepo := pgdb.NewPostgresDailyRepo(dbPool)
	timetrialRepo := pgdb.NewPostgresTimetrialRepo(dbPool)
	challengeService := service.NewChallengeGenService(musicProvider, genreRepo, dailyRepo, timetrialRepo)

	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("0 0 * * *", func() {
		slog.Info("Running daily challenge generation", "country", "spain")
		if err := challengeService.GenerateDailyChallenge("spain"); err != nil {
			slog.Error("Daily challenge generation failed", "error", err)
		} else {
			slog.Info("Daily challenge generation succeeded", "country", "spain")
		}
		slog.Info("Running inverse daily challenge generation", "country", "spain")
		if err := challengeService.GenerateInverseChallenge("spain"); err != nil {
			slog.Error("Inverse daily challenge generation failed", "error", err)
		} else {
			slog.Info("Inverse daily challenge generation succeeded", "country", "spain")
		}
	})
	if err != nil {
		slog.Error("Failed to schedule daily challenge job", "error", err)
		os.Exit(1)
	}

	c.Start()
	slog.Info("Daily challenge cron job scheduled", "schedule", "0 0 * * * (UTC)")
}

// runServer starts the HTTP server and blocks until an OS signal triggers a graceful shutdown.
func runServer(srv *http.Server, port string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Server listening", "port", port)
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

// healthHandler responds with a simple JSON status check.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		slog.Error("Failed to write health check response", "error", err)
	}
}
