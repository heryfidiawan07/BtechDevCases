package dto

type TransferRequest struct {
	Recipient string `json:"recipient" validate:"required,email,max=254"`
	Amount    string `json:"amount" validate:"required"`
	Notes     string `json:"notes" validate:"max=280"`
}

type TransferResponse struct {
	ID             int64  `json:"id"`
	FromUserID     int64  `json:"fromUserId"`
	ToUserID       int64  `json:"toUserId"`
	SenderEmail    string `json:"senderEmail"`
	RecipientEmail string `json:"recipientEmail"`
	Amount         string `json:"amount"`
	Notes          string `json:"notes"`
	Direction      string `json:"direction"`
	CreatedAt      string `json:"createdAt"`
}

type WalletResponse struct {
	Balance   string             `json:"balance"`
	Transfers []TransferResponse `json:"transfers"`
}
