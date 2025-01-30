package request_model

import (
	"github.com/google/uuid"
	"time"
)

type Review struct {
	id         uuid.UUID
	user_id    uuid.UUID
	product_id uuid.UUID
	rating     int
	comment    string
	is_deleted bool
	is_edited  bool
	created_at time.Time
	updated_at time.Time
}

type ReviewEditParams struct {
	id      uuid.UUID
	rating  int
	comment string
}

type ReviewDeleteParams struct {
	id uuid.UUID
}
