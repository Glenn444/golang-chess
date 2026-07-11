package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/docs"
	"github.com/Glenn444/golang-chess/internal/api"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"

)

// @title           Chess Game API
// @version         1.0.0
// @description     Multiplayer chess game server with WebSocket real-time gameplay, chat, and voice signalling.
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey Bearer
// @in              header
// @name            Authorization
// @description     JWT Bearer token. Prefix with "Bearer ".

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("error loading the config, ", err)
	}
	docs.SwaggerInfo.Host = cfg.PUBLIC_HOST
	dbConfig, err := pgxpool.ParseConfig(cfg.DB_URL)
	if err != nil {
		log.Fatal("cannot parse db config: ", err)
	}

	dbConfig.MaxConns = 10
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatal("cannot create db connection pool: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(pool)
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("failed to set up server: ", err)
	}

	// Start the server in a goroutine so we can listen for shutdown signals.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "address", cfg.ServerAddress)
		if err := server.Start(cfg.ServerAddress); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for interrupt signal or server error.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutting down", "signal", sig.String())
	case err := <-errCh:
		slog.Error("server error", "err", err)
	}

	// Graceful shutdown with a deadline.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced to shutdown", "err", err)
	}

	slog.Info("server stopped")
}
