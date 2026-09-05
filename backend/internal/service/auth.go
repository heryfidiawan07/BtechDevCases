package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/repository"
	"btechdevcases/internal/security"
)

type AuthService struct {
	users          *repository.UserRepository
	hasher         *security.PasswordHasher
	tokens         *security.JWTManager
	initialBalance decimal.Decimal
}

func NewAuthService(
	users *repository.UserRepository,
	hasher *security.PasswordHasher,
	tokens *security.JWTManager,
	initialBalance string,
) (*AuthService, error) {
	balance, err := decimal.NewFromString(initialBalance)
	if err != nil {
		return nil, fmt.Errorf("initial balance: %w", err)
	}

	return &AuthService{
		users:          users,
		hasher:         hasher,
		tokens:         tokens,
		initialBalance: balance,
	}, nil
}

type AuthResult struct {
	User      *domain.User
	Token     string
	ExpiresAt time.Time
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	email = normalizeEmail(email)

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Balance:      s.initialBalance,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.issue(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.issue(user)
}

func (s *AuthService) Refresh(userID int64, email string) (*AuthResult, error) {
	return s.issue(&domain.User{ID: userID, Email: email})
}

func (s *AuthService) Me(ctx context.Context, userID int64) (*domain.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *AuthService) issue(user *domain.User) (*AuthResult, error) {
	token, expiresAt, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func WelcomeMessage(email string) string {
	return fmt.Sprintf("Hello %s, welcome back", email)
}
