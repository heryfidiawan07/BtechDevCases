package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	var lastErr error
	for attempt := 1; attempt <= 20; attempt++ {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}

		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return pool, nil
		}

		pool.Close()
		lastErr = err
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("connect postgres: %w", lastErr)
}
