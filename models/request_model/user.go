package request_model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	id         uuid.UUID
	name       string
	email      string
	phone      int
	password   string
	is_blocked bool
	created_at time.Time
	updated_at time.Time
}

type UserSignUpParams struct {
	name     string
	email    string
	phone    int
	password string
}

type UserLoginParams struct {
	name     string
	email    string
	phone    int
	password string
}

type UserSignUpOTPParams struct {
	id  uuid.UUID
	otp int
}

type UserForgotPasswordParams struct {
	id uuid.UUID
}

type UserBlockParams struct {
	id uuid.UUID
}

type UserUnblockParams struct {
	id uuid.UUID
}
