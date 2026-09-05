package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/repository"
)

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,128}$`)

type WalletService struct {
	users     *repository.UserRepository
	transfers *repository.TransferRepository
}

func NewWalletService(users *repository.UserRepository, transfers *repository.TransferRepository) *WalletService {
	return &WalletService{users: users, transfers: transfers}
}

func (s *WalletService) GetWallet(ctx context.Context, userID int64) (*domain.User, []domain.Transfer, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	items, err := s.transfers.ListForUser(ctx, userID, 50)
	if err != nil {
		return nil, nil, err
	}

	return user, items, nil
}

func (s *WalletService) Transfer(
	ctx context.Context,
	fromUserID int64,
	recipientEmail string,
	amountRaw string,
	notes string,
	idempotencyKey string,
) (*domain.Transfer, error) {
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return nil, domain.ErrIdempotencyKeyRequired
	}
	if !idempotencyKeyPattern.MatchString(key) {
		return nil, domain.ErrIdempotencyKeyInvalid
	}

	amount, err := ParseAmount(amountRaw)
	if err != nil {
		return nil, err
	}

	recipientEmail = normalizeEmail(recipientEmail)
	notes = strings.TrimSpace(notes)

	existing, err := s.transfers.GetByIdempotencyKey(ctx, fromUserID, key)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	return s.transfers.ExecuteTransfer(ctx, repository.TransferParams{
		FromUserID:     fromUserID,
		RecipientEmail: recipientEmail,
		Amount:         amount,
		Notes:          notes,
		IdempotencyKey: key,
	})
}
