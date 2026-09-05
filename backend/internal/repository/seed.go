package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/security"
)

type demoAccount struct {
	Email    string
	Password string
}

var demoAccounts = []demoAccount{
	{Email: "fidiawan07@gmail.com", Password: "12345678"},
	{Email: "heryfidiawan07@gmail.com", Password: "12345678"},
}

func SeedDemoUsers(ctx context.Context, users *UserRepository, hasher *security.PasswordHasher, initialBalance string) error {
	balance, err := decimal.NewFromString(initialBalance)
	if err != nil {
		return err
	}

	for _, account := range demoAccounts {
		existing, err := users.GetByEmail(ctx, account.Email)
		if err == nil {
			if existing.Balance.LessThan(balance) {
				if err := users.SetBalance(ctx, existing.ID, balance); err != nil {
					return err
				}
				slog.Info("updated demo user balance", "email", account.Email, "balance", balance.StringFixed(2))
			}
			continue
		}
		if !errors.Is(err, domain.ErrUserNotFound) {
			return err
		}

		hash, err := hasher.Hash(account.Password)
		if err != nil {
			return err
		}

		user := &domain.User{
			Email:        account.Email,
			PasswordHash: hash,
			Balance:      balance,
		}
		if err := users.Create(ctx, user); err != nil && !errors.Is(err, domain.ErrEmailTaken) {
			return err
		}

		slog.Info("seeded demo user", "email", account.Email)
	}

	return nil
}
