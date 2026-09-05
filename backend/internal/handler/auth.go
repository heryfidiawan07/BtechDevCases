package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"btechdevcases/internal/dto"
	"btechdevcases/internal/middleware"
	"btechdevcases/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := bindJSON(c, &req); err != nil {
		return writeError(c, err)
	}

	result, err := h.auth.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return writeError(c, err)
	}

	return writeData(c, fiber.StatusCreated, toTokenResponse(result))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		return writeError(c, err)
	}

	result, err := h.auth.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return writeError(c, err)
	}

	return writeData(c, fiber.StatusOK, toTokenResponse(result))
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	principal, err := middleware.PrincipalFrom(c)
	if err != nil {
		return writeError(c, err)
	}

	result, err := h.auth.Refresh(principal.ID, principal.Email)
	if err != nil {
		return writeError(c, err)
	}

	return writeData(c, fiber.StatusOK, toTokenResponse(result))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return writeData(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	principal, err := middleware.PrincipalFrom(c)
	if err != nil {
		return writeError(c, err)
	}

	user, err := h.auth.Me(c.Context(), principal.ID)
	if err != nil {
		return writeError(c, err)
	}

	return writeData(c, fiber.StatusOK, dto.MeResponse{
		ID:      user.ID,
		Email:   user.Email,
		Message: service.WelcomeMessage(user.Email),
	})
}

func toTokenResponse(result *service.AuthResult) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken: result.Token,
		ExpiresAt:   result.ExpiresAt.UTC().Format(time.RFC3339),
		User: dto.UserPublic{
			ID:    result.User.ID,
			Email: result.User.Email,
		},
	}
}
