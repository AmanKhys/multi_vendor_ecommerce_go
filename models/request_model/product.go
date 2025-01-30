package request_model

import (
	"github.com/google/uuid"
	"time"
)

type Product struct {
	id          uuid.UUID
	name        string
	description string
	price       float64
	stock       int
	seller_id   uuid.UUID
	category_id uuid.UUID
	is_deleted  bool
	created_at  time.Time
	updated_at  time.Time
}

type ProductAddParams struct {
	name        string
	description string
	price       float64
	stock       int
	seller_id   uuid.UUID
	category_id uuid.UUID
}

type ProductEditParams struct {
	id          uuid.UUID
	name        string
	description string
	price       float64
	stock       int
	category_id uuid.UUID
}

type ProductDeleteParams struct {
	id         uuid.UUID
	is_deleted bool
}
