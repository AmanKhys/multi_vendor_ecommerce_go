package request_model

import (
	"github.com/google/uuid"
	"time"
)

type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryAddParams struct {
	Name string `json:"name"`
}

type CategoryEditParams struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CategoryDeleteParams struct {
	ID uuid.UUID `json:"id"`
}

