package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"btechdevcases/internal/config"
	"btechdevcases/internal/migrate"
	"btechdevcases/internal/repository"
	"btechdevcases/internal/security"
	"btechdevcases/internal/server"
	"btechdevcases/internal/service"
)

func TestAuthAndTransferFlow(t *testing.T) {
	app := newTestApp(t)

	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	emailA := "flow-a-" + suffix + "@example.com"
	emailB := "flow-b-" + suffix + "@example.com"

	register := postJSON(t, app, "/api/v1/auth/register", nil, map[string]string{
		"email":           emailA,
		"password":        "Secret123",
		"confirmPassword": "Secret123",
	})
	if register.StatusCode != fiber.StatusCreated {
		t.Fatalf("register status %d body %s", register.StatusCode, register.Body)
	}

	loginB := postJSON(t, app, "/api/v1/auth/register", nil, map[string]string{
		"email":           emailB,
		"password":        "Secret123",
		"confirmPassword": "Secret123",
	})
	if loginB.StatusCode != fiber.StatusCreated {
		t.Fatalf("register b status %d body %s", loginB.StatusCode, loginB.Body)
	}

	tokenA := tokenFrom(t, register.Body)
	tokenB := tokenFrom(t, loginB.Body)

	me := getJSON(t, app, "/api/v1/me", tokenA)
	if me.StatusCode != fiber.StatusOK {
		t.Fatalf("me status %d", me.StatusCode)
	}
	if !bytes.Contains(me.Body, []byte(fmt.Sprintf("Hello %s, welcome back", emailA))) {
		t.Fatalf("unexpected me body %s", me.Body)
	}

	key := "550e8400-e29b-41d4-a716-446655440000"
	first := postJSON(t, app, "/api/v1/wallet/transfers", map[string]string{
		"Authorization":   "Bearer " + tokenA,
		"Idempotency-Key": key,
	}, map[string]string{
		"recipient": emailB,
		"amount":    "25.50",
		"notes":     "camp supplies",
	})
	if first.StatusCode != fiber.StatusCreated {
		t.Fatalf("transfer status %d body %s", first.StatusCode, first.Body)
	}

	replay := postJSON(t, app, "/api/v1/wallet/transfers", map[string]string{
		"Authorization":   "Bearer " + tokenA,
		"Idempotency-Key": key,
	}, map[string]string{
		"recipient": emailB,
		"amount":    "25.50",
		"notes":     "camp supplies",
	})
	if replay.StatusCode != fiber.StatusCreated {
		t.Fatalf("replay status %d body %s", replay.StatusCode, replay.Body)
	}

	walletA := getJSON(t, app, "/api/v1/wallet", tokenA)
	walletB := getJSON(t, app, "/api/v1/wallet", tokenB)
	if !bytes.Contains(walletA.Body, []byte(`"balance":"9974.50"`)) {
		t.Fatalf("sender wallet %s", walletA.Body)
	}
	if !bytes.Contains(walletB.Body, []byte(`"balance":"10025.50"`)) {
		t.Fatalf("recipient wallet %s", walletB.Body)
	}

	poor := postJSON(t, app, "/api/v1/wallet/transfers", map[string]string{
		"Authorization":   "Bearer " + tokenA,
		"Idempotency-Key": "550e8400-e29b-41d4-a716-446655440099",
	}, map[string]string{
		"recipient": emailB,
		"amount":    "999999",
		"notes":     "too much",
	})
	if poor.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("insufficient status %d body %s", poor.StatusCode, poor.Body)
	}
}

type httpResult struct {
	StatusCode int
	Body       []byte
}

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	t.Setenv("JWT_SECRET", "integration-test-secret-key-32chars")
	t.Setenv("INITIAL_BALANCE", "10000")
	t.Setenv("CORS_ORIGIN", "http://localhost:3000")
	t.Setenv("HTTP_PORT", "8080")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DatabaseURL = databaseURL

	ctx := context.Background()
	if err := migrate.Up(databaseURL); err != nil {
		t.Fatal(err)
	}

	pool, err := repository.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	users := repository.NewUserRepository(pool)
	transfers := repository.NewTransferRepository(pool)
	hasher := security.NewPasswordHasher()
	tokens := security.NewJWTManager(cfg.JWTSecret, 15*time.Minute)
	authService, err := service.NewAuthService(users, hasher, tokens, cfg.InitialBalance)
	if err != nil {
		t.Fatal(err)
	}

	return server.New(server.Dependencies{
		Config: cfg,
		Pool:   pool,
		Auth:   authService,
		Wallet: service.NewWalletService(users, transfers),
		Tokens: tokens,
	})
}

func postJSON(t *testing.T, app *fiber.App, path string, headers map[string]string, payload any) httpResult {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return httpResult{StatusCode: resp.StatusCode, Body: raw}
}

func getJSON(t *testing.T, app *fiber.App, path, token string) httpResult {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return httpResult{StatusCode: resp.StatusCode, Body: raw}
}

func tokenFrom(t *testing.T, body []byte) string {
	t.Helper()

	var parsed struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Data.AccessToken == "" {
		t.Fatalf("missing token in %s", body)
	}

	return parsed.Data.AccessToken
}
