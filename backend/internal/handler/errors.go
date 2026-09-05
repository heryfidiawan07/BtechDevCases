package handler

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/dto"
)

func writeData(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(dto.APIResponse{Data: data})
}

func writeError(c *fiber.Ctx, err error) error {
	code, status, message, details := mapError(err)
	if status >= 500 {
		slog.Error("request failed", "error", err, "path", c.Path())
	}

	return c.Status(status).JSON(dto.APIError{
		Error: dto.ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func mapError(err error) (code string, status int, message string, details map[string]string) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		details = validationDetails(validationErrs)
		return "VALIDATION_ERROR", fiber.StatusBadRequest, "Validation failed", details
	}

	var fundsErr *domain.InsufficientFundsError
	if errors.As(err, &fundsErr) {
		return "INSUFFICIENT_FUNDS", fiber.StatusUnprocessableEntity, fundsMessage(fundsErr), nil
	}

	switch {
	case errors.Is(err, domain.ErrInvalidBody):
		return "INVALID_BODY", fiber.StatusBadRequest, "Invalid request body", nil
	case errors.Is(err, domain.ErrValidation):
		return "VALIDATION_ERROR", fiber.StatusBadRequest, err.Error(), nil
	case errors.Is(err, domain.ErrEmailTaken):
		return "EMAIL_TAKEN", fiber.StatusConflict, "Email is already registered", nil
	case errors.Is(err, domain.ErrInvalidCredentials):
		return "INVALID_CREDENTIALS", fiber.StatusUnauthorized, "Invalid email or password", nil
	case errors.Is(err, domain.ErrUnauthorized):
		return "UNAUTHORIZED", fiber.StatusUnauthorized, "Unauthorized", nil
	case errors.Is(err, domain.ErrRecipientNotFound):
		return "RECIPIENT_NOT_FOUND", fiber.StatusNotFound, "Recipient not found", nil
	case errors.Is(err, domain.ErrSelfTransfer):
		return "SELF_TRANSFER", fiber.StatusBadRequest, "You cannot transfer to yourself", nil
	case errors.Is(err, domain.ErrInsufficientFunds):
		return "INSUFFICIENT_FUNDS", fiber.StatusUnprocessableEntity, "Insufficient funds. The requested amount exceeds your available balance.", nil
	case errors.Is(err, domain.ErrInvalidAmount):
		return "INVALID_AMOUNT", fiber.StatusBadRequest, "Amount must be a positive number with an optional comma and up to 2 decimal places", nil
	case errors.Is(err, domain.ErrIdempotencyKeyRequired):
		return "IDEMPOTENCY_KEY_REQUIRED", fiber.StatusBadRequest, "Idempotency-Key header is required", nil
	case errors.Is(err, domain.ErrIdempotencyKeyInvalid):
		return "IDEMPOTENCY_KEY_INVALID", fiber.StatusBadRequest, "Idempotency-Key is invalid", nil
	case errors.Is(err, domain.ErrUserNotFound):
		return "NOT_FOUND", fiber.StatusNotFound, "Resource not found", nil
	default:
		return "INTERNAL_ERROR", fiber.StatusInternalServerError, "Internal server error", nil
	}
}

func fundsMessage(err *domain.InsufficientFundsError) string {
	return "Insufficient funds. Available " + err.Available.StringFixed(2) + ", requested " + err.Requested.StringFixed(2) + "."
}

func validationDetails(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(errs))
	for _, fieldErr := range errs {
		details[fieldErr.Field()] = validationMessage(fieldErr)
	}

	return details
}

func validationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email"
	case "min":
		return "Is too short"
	case "max":
		return "Is too long"
	case "eqfield":
		return "Must match password"
	default:
		return "Is invalid"
	}
}

func bindJSON(c *fiber.Ctx, dest any) error {
	if err := c.BodyParser(dest); err != nil {
		return domain.ErrInvalidBody
	}

	return validate.Struct(dest)
}
