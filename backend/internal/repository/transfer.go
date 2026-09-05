package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
)

type TransferParams struct {
	FromUserID     int64
	RecipientEmail string
	Amount         decimal.Decimal
	Notes          string
	IdempotencyKey string
}

type TransferRepository struct {
	pool *pgxpool.Pool
}

func NewTransferRepository(pool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{pool: pool}
}

func (r *TransferRepository) GetByIdempotencyKey(ctx context.Context, fromUserID int64, key string) (*domain.Transfer, error) {
	const q = transferSelect + `
		WHERE t.from_user_id = $1 AND t.idempotency_key = $2
	`

	return scanTransfer(r.pool.QueryRow(ctx, q, fromUserID, key))
}

func (r *TransferRepository) ListForUser(ctx context.Context, userID int64, limit int) ([]domain.Transfer, error) {
	const q = transferSelect + `
		WHERE t.from_user_id = $1 OR t.to_user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list transfers: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Transfer, 0)
	for rows.Next() {
		item, err := scanTransferRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list transfers rows: %w", err)
	}

	return items, nil
}

func (r *TransferRepository) ExecuteTransfer(ctx context.Context, params TransferParams) (*domain.Transfer, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transfer tx: %w", err)
	}
	defer tx.Rollback(ctx)

	existing, err := scanTransfer(tx.QueryRow(ctx, transferSelect+`
		WHERE t.from_user_id = $1 AND t.idempotency_key = $2
	`, params.FromUserID, params.IdempotencyKey))
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	recipient, err := scanUser(tx.QueryRow(ctx, userSelect+" WHERE email = $1", params.RecipientEmail))
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrRecipientNotFound
	}
	if err != nil {
		return nil, err
	}
	if recipient.ID == params.FromUserID {
		return nil, domain.ErrSelfTransfer
	}

	firstID, secondID := params.FromUserID, recipient.ID
	if secondID < firstID {
		firstID, secondID = secondID, firstID
	}

	first, err := lockUser(ctx, tx, firstID)
	if err != nil {
		return nil, err
	}
	second, err := lockUser(ctx, tx, secondID)
	if err != nil {
		return nil, err
	}

	sender, counterparty := first, second
	if first.ID != params.FromUserID {
		sender, counterparty = second, first
	}

	if sender.Balance.LessThan(params.Amount) {
		return nil, &domain.InsufficientFundsError{Available: sender.Balance, Requested: params.Amount}
	}

	sender.Balance = sender.Balance.Sub(params.Amount)
	counterparty.Balance = counterparty.Balance.Add(params.Amount)

	if err := updateBalance(ctx, tx, sender.ID, sender.Balance); err != nil {
		return nil, err
	}
	if err := updateBalance(ctx, tx, counterparty.ID, counterparty.Balance); err != nil {
		return nil, err
	}

	const insertQ = `
		INSERT INTO transfers (from_user_id, to_user_id, amount, notes, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	transfer := &domain.Transfer{
		FromUserID:     params.FromUserID,
		ToUserID:       recipient.ID,
		SenderEmail:    sender.Email,
		RecipientEmail: recipient.Email,
		Amount:         params.Amount,
		Notes:          params.Notes,
		IdempotencyKey: params.IdempotencyKey,
	}

	err = tx.QueryRow(ctx, insertQ, params.FromUserID, recipient.ID, params.Amount.StringFixed(2), params.Notes, params.IdempotencyKey).
		Scan(&transfer.ID, &transfer.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			replay, replayErr := scanTransfer(tx.QueryRow(ctx, transferSelect+`
				WHERE t.from_user_id = $1 AND t.idempotency_key = $2
			`, params.FromUserID, params.IdempotencyKey))
			if replayErr != nil {
				return nil, fmt.Errorf("replay idempotent transfer: %w", replayErr)
			}

			return replay, nil
		}

		return nil, fmt.Errorf("insert transfer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transfer: %w", err)
	}

	return transfer, nil
}

func lockUser(ctx context.Context, tx pgx.Tx, id int64) (*domain.User, error) {
	return scanUser(tx.QueryRow(ctx, userSelect+" WHERE id = $1 FOR UPDATE", id))
}

func updateBalance(ctx context.Context, tx pgx.Tx, id int64, balance decimal.Decimal) error {
	const q = `
		UPDATE users
		SET balance = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(ctx, q, id, balance.StringFixed(2))
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}

const transferSelect = `
	SELECT
		t.id,
		t.from_user_id,
		t.to_user_id,
		sender.email,
		recipient.email,
		t.amount::text,
		t.notes,
		t.idempotency_key,
		t.created_at
	FROM transfers t
	JOIN users sender ON sender.id = t.from_user_id
	JOIN users recipient ON recipient.id = t.to_user_id
`

func scanTransfer(row pgx.Row) (*domain.Transfer, error) {
	item, err := scanTransferDest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	return item, err
}

func scanTransferRow(rows pgx.Rows) (*domain.Transfer, error) {
	return scanTransferDest(rows)
}

func scanTransferDest(row interface{ Scan(dest ...any) error }) (*domain.Transfer, error) {
	var (
		item      domain.Transfer
		amountRaw string
	)

	err := row.Scan(
		&item.ID,
		&item.FromUserID,
		&item.ToUserID,
		&item.SenderEmail,
		&item.RecipientEmail,
		&amountRaw,
		&item.Notes,
		&item.IdempotencyKey,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	amount, err := decimal.NewFromString(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("parse transfer amount: %w", err)
	}
	item.Amount = amount

	return &item, nil
}
