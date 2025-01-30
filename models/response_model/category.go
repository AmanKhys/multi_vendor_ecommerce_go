package response_model

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
