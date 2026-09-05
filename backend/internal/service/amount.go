package service

import (
	"strings"

	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
)

var maxAmount = decimal.RequireFromString("99999999999999.99")

func ParseAmount(raw string) (decimal.Decimal, error) {
	normalized := normalizeAmount(strings.TrimSpace(raw))
	value, err := decimal.NewFromString(normalized)
	if err != nil {
		return decimal.Zero, domain.ErrInvalidAmount
	}
	if !value.GreaterThan(decimal.Zero) {
		return decimal.Zero, domain.ErrInvalidAmount
	}
	if value.Exponent() < -2 {
		return decimal.Zero, domain.ErrInvalidAmount
	}
	if value.GreaterThan(maxAmount) {
		return decimal.Zero, domain.ErrInvalidAmount
	}

	return value, nil
}

func normalizeAmount(raw string) string {
	if !strings.Contains(raw, ",") {
		return raw
	}

	withoutThousands := strings.ReplaceAll(raw, ".", "")
	return strings.TrimRight(strings.ReplaceAll(withoutThousands, ",", "."), ".")
}
