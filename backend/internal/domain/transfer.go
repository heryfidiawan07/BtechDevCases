package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transfer struct {
	ID             int64
	FromUserID     int64
	ToUserID       int64
	SenderEmail    string
	RecipientEmail string
	Amount         decimal.Decimal
	Notes          string
	IdempotencyKey string
	CreatedAt      time.Time
}
