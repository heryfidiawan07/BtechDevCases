package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (email, password_hash, balance)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, q, user.Email, user.PasswordHash, user.Balance.StringFixed(2)).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}

		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, userSelect+" WHERE email = $1", strings.ToLower(email)))
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx, userSelect+" WHERE id = $1", id))
}

func (r *UserRepository) SetBalance(ctx context.Context, userID int64, balance decimal.Decimal) error {
	const q = `
		UPDATE users
		SET balance = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, q, userID, balance.StringFixed(2))
	if err != nil {
		return fmt.Errorf("set balance: %w", err)
	}

	return nil
}

const userSelect = `
	SELECT id, email, password_hash, balance::text, created_at, updated_at
	FROM users
`

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		user       domain.User
		balanceRaw string
	)

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &balanceRaw, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	balance, err := decimal.NewFromString(balanceRaw)
	if err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	}
	user.Balance = balance

	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
