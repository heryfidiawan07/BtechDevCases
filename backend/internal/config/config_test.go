package config

import (
	"testing"
	"time"
)

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable")
	t.Setenv("JWT_SECRET", "too-short")
	t.Setenv("JWT_TTL", "15m")
	t.Setenv("INITIAL_BALANCE", "10000")

	if _, err := Load(); err == nil {
		t.Fatal("expected JWT_SECRET validation error")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable")
	t.Setenv("JWT_SECRET", "change-me-to-a-long-random-secret-key")
	t.Setenv("JWT_TTL", "")
	t.Setenv("INITIAL_BALANCE", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("CORS_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWTTTL != 15*time.Minute {
		t.Fatalf("ttl %s", cfg.JWTTTL)
	}
	if cfg.InitialBalance != "100000" {
		t.Fatalf("balance %s", cfg.InitialBalance)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("port %s", cfg.HTTPPort)
	}
}
