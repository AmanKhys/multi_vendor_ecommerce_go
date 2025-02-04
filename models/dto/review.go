package dto

import (
	"github.com/google/uuid"
	"time"
)

type Review struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	IsDeleted bool      `json:"is_deleted"`
	IsEdited  bool      `json:"is_edited"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewEditParams struct {
	ID      uuid.UUID `json:"id"`
	Rating  int       `json:"rating"`
	Comment string    `json:"comment"`
}

type ReviewDeleteParams struct {
	ID uuid.UUID `json:"id"`
}
