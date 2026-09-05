package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"btechdevcases/internal/config"
	"btechdevcases/internal/migrate"
	"btechdevcases/internal/repository"
	"btechdevcases/internal/security"
	"btechdevcases/internal/server"
	"btechdevcases/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := migrate.UpWithRetry(cfg.DatabaseURL, 20, time.Second); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	users := repository.NewUserRepository(pool)
	transfers := repository.NewTransferRepository(pool)
	hasher := security.NewPasswordHasher()
	tokens := security.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)

	if err := repository.SeedDemoUsers(ctx, users, hasher, cfg.InitialBalance); err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	authService, err := service.NewAuthService(users, hasher, tokens, cfg.InitialBalance)
	if err != nil {
		return err
	}

	app := server.New(server.Dependencies{
		Config: cfg,
		Pool:   pool,
		Auth:   authService,
		Wallet: service.NewWalletService(users, transfers),
		Tokens: tokens,
	})

	errCh := make(chan error, 1)
	go func() {
		addr := ":" + cfg.HTTPPort
		slog.Info("listening", "addr", addr)
		if err := app.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return app.ShutdownWithContext(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
