package entity

import (
	"github.com/google/uuid"
	"time"
)

type Seller struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	About     string    `json:"about"`
	Email     string    `json:"email"`
	Phone     int       `json:"phone"`
	GstNo     string    `json:"gst_no"`
	Password  string    `json:"password"`
	IsBlocked bool      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
