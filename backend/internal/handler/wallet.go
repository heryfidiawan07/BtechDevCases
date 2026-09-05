package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"btechdevcases/internal/domain"
	"btechdevcases/internal/dto"
	"btechdevcases/internal/middleware"
	"btechdevcases/internal/service"
)

type WalletHandler struct {
	wallet *service.WalletService
}

func NewWalletHandler(wallet *service.WalletService) *WalletHandler {
	return &WalletHandler{wallet: wallet}
}

func (h *WalletHandler) Get(c *fiber.Ctx) error {
	principal, err := middleware.PrincipalFrom(c)
	if err != nil {
		return writeError(c, err)
	}

	user, transfers, err := h.wallet.GetWallet(c.Context(), principal.ID)
	if err != nil {
		return writeError(c, err)
	}

	items := make([]dto.TransferResponse, 0, len(transfers))
	for _, item := range transfers {
		items = append(items, toTransferResponse(item, principal.ID))
	}

	return writeData(c, fiber.StatusOK, dto.WalletResponse{
		Balance:   user.Balance.StringFixed(2),
		Transfers: items,
	})
}

func (h *WalletHandler) Transfer(c *fiber.Ctx) error {
	principal, err := middleware.PrincipalFrom(c)
	if err != nil {
		return writeError(c, err)
	}

	var req dto.TransferRequest
	if err := bindJSON(c, &req); err != nil {
		return writeError(c, err)
	}

	item, err := h.wallet.Transfer(
		c.Context(),
		principal.ID,
		req.Recipient,
		req.Amount,
		req.Notes,
		c.Get("Idempotency-Key"),
	)
	if err != nil {
		return writeError(c, err)
	}

	return writeData(c, fiber.StatusCreated, toTransferResponse(*item, principal.ID))
}

func toTransferResponse(item domain.Transfer, viewerID int64) dto.TransferResponse {
	direction := "in"
	if item.FromUserID == viewerID {
		direction = "out"
	}

	return dto.TransferResponse{
		ID:             item.ID,
		FromUserID:     item.FromUserID,
		ToUserID:       item.ToUserID,
		SenderEmail:    item.SenderEmail,
		RecipientEmail: item.RecipientEmail,
		Amount:         item.Amount.StringFixed(2),
		Notes:          item.Notes,
		Direction:      direction,
		CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
	}
}
