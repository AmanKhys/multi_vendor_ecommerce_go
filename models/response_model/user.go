package response_model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	id         uuid.UUID
	name       string
	email      string
	phone      int
	password   string
	is_blocked bool
	created_at time.Time
	updated_at time.Time
}

type UserProfileRes struct {
	id   uuid.UUID
	name string
}

type UserAdminRes struct {
	id         uuid.UUID
	name       string
	email      string
	phone      string
	is_blocked string
	created_at time.Time
	updated_at time.Time
}
