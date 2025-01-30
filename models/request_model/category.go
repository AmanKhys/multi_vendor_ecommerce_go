package request_model

import (
	"github.com/google/uuid"
	"time"
)

type Category struct {
	id         uuid.UUID
	name       string
	is_deleted bool
	created_at time.Time
	updated_at time.Time
}

type CateogryAddParams struct {
	name string
}

type CategoryEditParams struct {
	id   uuid.UUID
	name bool
}

type CateogryDeleteParams struct {
	id uuid.UUID
}
