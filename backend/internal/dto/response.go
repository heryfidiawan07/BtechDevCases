package dto

type APIResponse struct {
	Data any `json:"data,omitempty"`
}

type APIError struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type UserPublic struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type TokenResponse struct {
	AccessToken string     `json:"accessToken"`
	ExpiresAt   string     `json:"expiresAt"`
	User        UserPublic `json:"user"`
}

type MeResponse struct {
	ID      int64  `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}
