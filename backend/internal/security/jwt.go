package security

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"btechdevcases/internal/domain"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *JWTManager) Issue(userID int64, email string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *JWTManager) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}

func UserIDFromClaims(claims *Claims) (int64, error) {
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return id, nil
}
