-- name: EditUserByID :one
update users
set name = $2, phone = $3, updated_at = current_timestamp
where id = $1
returning *;

-- name: EditSellerByID :one
update users
set name = $2, about = $3, phone = $4, updated_at = current_timestamp
where id = $1
returning *;

-- name: AddUser :one
INSERT INTO users
(name, email, phone, password, role)
VALUES ($1, $2, $3, $4, 'user')
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified;

-- name: AddAndVerifyUser :one
insert into users
(name, email, password, role, email_verified, user_verified, updated_at)
values  ($1, $2, $3, 'user', true, true, current_timestamp)
returning id, name, email, role, is_blocked, email_verified, user_verified;

-- name: AddSeller :one
INSERT INTO users
(name, email, phone, password, role, gst_no, about)
VALUES ($1, $2, $3, $4, 'seller', $5, $6)
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about;

-- name: VerifyUserByID :one
UPDATE users
SET email_verified = true, user_verified = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified;

-- name: VerifySellerEmailByID :one
UPDATE users
SET email_verified = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified;

-- name: VerifySellerByID :one
update users
set user_verified = true, updated_at = current_timestamp
where id = $1
returning id, name, email, phone, role, is_blocked, email_verified, user_verified;

-- name: GetAllUsers :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about FROM users;

-- name: GetAllUsersByRoleSeller :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, about, gst_no FROM users
WHERE role = $1;

-- name: GetAllUsersByRoleUser :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified FROM users
WHERE role = $1;

-- name: GetUserById :one
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about FROM users
WHERE id = $1;


-- name: GetUserWithPasswordByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByEmail :one
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about FROM users
WHERE email = $1;

-- name: BlockUserByID :one
UPDATE users
SET is_blocked = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked;

-- name: UnblockUserByID :one
UPDATE users
SET is_blocked = false, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked;


-- name: ChangePasswordByUserID :exec
update users
set password = $2
where id = $1;

-- name: ChangeNameByUserID :exec
update users
set name = $2
where id = $1;