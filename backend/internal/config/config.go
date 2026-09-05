package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	HTTPPort       string
	DatabaseURL    string
	JWTSecret      string
	JWTTTL         time.Duration
	InitialBalance string
	CORSOrigin     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	ttl, err := parseDuration(envOr("JWT_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("JWT_TTL: %w", err)
	}

	cfg := &Config{
		Env:            envOr("APP_ENV", "development"),
		HTTPPort:       envOr("HTTP_PORT", "8080"),
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:      strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTTTL:         ttl,
		InitialBalance: envOr("INITIAL_BALANCE", "100000"),
		CORSOrigin:     envOr("CORS_ORIGIN", "http://localhost:3000"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.JWTTTL <= 0 {
		return fmt.Errorf("JWT_TTL must be greater than 0")
	}
	if _, err := strconv.ParseFloat(c.InitialBalance, 64); err != nil {
		return fmt.Errorf("INITIAL_BALANCE must be a number")
	}
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}

	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func parseDuration(raw string) (time.Duration, error) {
	ttl, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}

	return ttl, nil
}
