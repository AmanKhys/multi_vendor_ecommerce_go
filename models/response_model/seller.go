package response_model

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

type SellerProfileRes struct {
	id         uuid.UUID
	name       string
	about      string
	email      string
	phone      int
	gst_no     string
	created_at time.Time
	updated_at time.Time
}

type SellerAdminRes struct {
	id         uuid.UUID
	name       string
	about      string
	email      string
	phone      int
	gst_no     string
	is_blocked bool
	created_at time.Time
	updated_at time.Time
}
