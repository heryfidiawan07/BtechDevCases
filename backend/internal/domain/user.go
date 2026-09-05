package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Balance      decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
