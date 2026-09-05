package handler

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err    error
		code   string
		status int
	}{
		{err: domain.ErrEmailTaken, code: "EMAIL_TAKEN", status: fiber.StatusConflict},
		{err: domain.ErrInvalidCredentials, code: "INVALID_CREDENTIALS", status: fiber.StatusUnauthorized},
		{err: domain.ErrInsufficientFunds, code: "INSUFFICIENT_FUNDS", status: fiber.StatusUnprocessableEntity},
		{err: domain.ErrSelfTransfer, code: "SELF_TRANSFER", status: fiber.StatusBadRequest},
		{err: domain.ErrRecipientNotFound, code: "RECIPIENT_NOT_FOUND", status: fiber.StatusNotFound},
		{err: domain.ErrIdempotencyKeyRequired, code: "IDEMPOTENCY_KEY_REQUIRED", status: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			code, status, message, _ := mapError(tt.err)
			if code != tt.code {
				t.Fatalf("code %s", code)
			}
			if status != tt.status {
				t.Fatalf("status %d", status)
			}
			if message == "" {
				t.Fatal("empty message")
			}
		})
	}
}

func TestMapInsufficientFundsDetails(t *testing.T) {
	t.Parallel()

	err := &domain.InsufficientFundsError{
		Available: decimal.RequireFromString("10000.00"),
		Requested: decimal.RequireFromString("50000.00"),
	}
	code, status, message, _ := mapError(err)
	if code != "INSUFFICIENT_FUNDS" {
		t.Fatalf("code %s", code)
	}
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status %d", status)
	}
	if message != "Insufficient funds. Available 10000.00, requested 50000.00." {
		t.Fatalf("message %q", message)
	}
}
