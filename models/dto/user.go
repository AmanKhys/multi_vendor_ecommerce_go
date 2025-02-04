package dto

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
	Role     string `json:"role"`
	GstNo    string `json:"gst_no"`
	About    string `json:"about"`
}

type UserLoginParams struct {
	Email    string `json:"email"`
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

// response part
type UserProfileRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserAdminRes struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	IsBlocked string    `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
