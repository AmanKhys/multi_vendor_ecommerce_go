-- name: AddUser :one
INSERT INTO users
(name, email, phone, password, role)
VALUES ($1, $2, $3, $4, 'user')
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, created_at, updated_at;

-- name: AddSeller :one
INSERT INTO users
(name, email, phone, password, role, gst_no, about)
VALUES ($1, $2, $3, $4, 'seller', $5, $6)
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at;

-- name: VerifyUserEmailByID :one
UPDATE users
SET email_verified = true, user_verified = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, created_at, updated_at;

-- name: VerifySellerEmailByID :one
UPDATE users
SET email_verified = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, created_at, updated_at;

-- name: VerifySellerUserByID :one
update users
set user_verified = true, updated_at = current_timestamp
where id = $1
returning id, name, email, phone, role, is_blocked, email_verified, user_verified, created_at, updated_at;

-- name: GetAllUsers :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at FROM users;

-- name: GetAllUsersByRole :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at FROM users
WHERE role = $1;

-- name: GetUserById :one
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at FROM users
WHERE id = $1;

-- name: GetUserWithPasswordByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByEmail :one
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at FROM users
WHERE email = $1;

-- name: GetUsersByRole :many
SELECT id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at FROM users
WHERE role = $1;

-- name: BlockUserByID :one
UPDATE users
SET is_blocked = true, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at;

-- name: UnblockUserByID :one
UPDATE users
SET is_blocked = false, updated_at = current_timestamp
WHERE id = $1
RETURNING id, name, email, phone, role, is_blocked, email_verified, user_verified, gst_no, about, created_at, updated_at;

-- name: GetOTPByUserID :one
SELECT * FROM login_otps
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;
