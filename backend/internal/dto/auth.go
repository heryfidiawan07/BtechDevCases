package dto

type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email,max=254"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
