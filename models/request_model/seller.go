package request_model

import (
	"github.com/google/uuid"
	"time"
)

type Seller struct {
	id         uuid.UUID
	name       string
	about      string
	email      string
	phone      int
	gst_no     string
	password   string
	is_blocked bool
	created_at time.Time
	updated_at time.Time
}

type SellerSignUpParams struct {
	name     string
	email    string
	phone    string
	gst_no   string
	password string
}

type SellerLoginParams struct {
	name     string
	email    string
	phone    string
	password string
}

type SellerSignUpOTPParams struct {
	id  uuid.UUID
	otp int
}

type SellerBlockParams struct {
	id uuid.UUID
}

type SellerUnblockParams struct {
	id uuid.UUID
}
