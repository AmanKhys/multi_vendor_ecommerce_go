package response_model

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
