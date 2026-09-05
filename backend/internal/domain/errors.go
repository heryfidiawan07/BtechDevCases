package domain

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidBody            = errors.New("invalid request body")
	ErrValidation             = errors.New("validation failed")
	ErrEmailTaken             = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrUserNotFound           = errors.New("user not found")
	ErrRecipientNotFound      = errors.New("recipient not found")
	ErrSelfTransfer           = errors.New("cannot transfer to yourself")
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrIdempotencyKeyRequired = errors.New("idempotency key required")
	ErrIdempotencyKeyInvalid  = errors.New("idempotency key invalid")
)

type InsufficientFundsError struct {
	Available decimal.Decimal
	Requested decimal.Decimal
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds: available %s, requested %s", e.Available.StringFixed(2), e.Requested.StringFixed(2))
}

func (e *InsufficientFundsError) Unwrap() error {
	return ErrInsufficientFunds
}
