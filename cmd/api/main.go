package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luca5Eckert/trama/internal/platform/config"
	"github.com/Luca5Eckert/trama/internal/platform/database"
	platformhttp "github.com/Luca5Eckert/trama/internal/platform/http"
	"github.com/Luca5Eckert/trama/internal/users"
	userpostgres "github.com/Luca5Eckert/trama/internal/users/infrastructure/postgres"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()
	pool, err := database.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	usersModule := users.NewModule(userpostgres.NewUserRepository(pool))
	router := platformhttp.NewRouter(logger, usersModule.RegisterRoutes)
	server := &http.Server{Addr: cfg.HTTPAddress, Handler: router, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		logger.Info("http server started", "address", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
