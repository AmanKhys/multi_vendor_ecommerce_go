package request_model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     int       `json:"phone"`
	Password  string    `json:"password"`
	IsBlocked bool      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSignUpParams struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    int    `json:"phone"`
	Password string `json:"password"`
}

type UserLoginParams struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    int    `json:"phone"`
	Password string `json:"password"`
}

type UserSignUpOTPParams struct {
	ID  uuid.UUID `json:"id"`
	OTP int       `json:"otp"`
}

type UserForgotPasswordParams struct {
	ID uuid.UUID `json:"id"`
}

type UserBlockParams struct {
	ID uuid.UUID `json:"id"`
}

type UserUnblockParams struct {
	ID uuid.UUID `json:"id"`
}

