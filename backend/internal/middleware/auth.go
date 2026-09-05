package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/security"
)

type Principal struct {
	ID    int64
	Email string
}

const principalKey = "principal"

func Authenticate(tokens *security.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := bearerToken(c.Get("Authorization"))
		if raw == "" {
			return unauthorized(c)
		}

		claims, err := tokens.Parse(raw)
		if err != nil {
			return unauthorized(c)
		}

		userID, err := security.UserIDFromClaims(claims)
		if err != nil {
			return unauthorized(c)
		}

		c.Locals(principalKey, Principal{ID: userID, Email: claims.Email})

		token, expiresAt, err := tokens.Issue(userID, claims.Email)
		if err == nil {
			c.Set("X-Access-Token", token)
			c.Set("X-Access-Token-Expires-At", expiresAt.UTC().Format(time.RFC3339))
		}

		return c.Next()
	}
}

func PrincipalFrom(c *fiber.Ctx) (Principal, error) {
	principal, ok := c.Locals(principalKey).(Principal)
	if !ok {
		return Principal{}, domain.ErrUnauthorized
	}

	return principal, nil
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Unauthorized",
		},
	})
}
