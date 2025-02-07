-- name: AddUser :one
insert into users
(name, email, phone, password, role, gst_no, about)
values ($1, $2, $3, $4, $5, $6, $7)
returning id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at;


-- name: GetAllUsers :many
select id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at from users;

-- name: GetAllUsersByRole :many
select id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at from users
where role = $1;

-- name: GetUserById :one
select id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at from users
where id = $1;

-- name: GetUserWithPasswordByID :one
select * from users
where id = $1;

-- name: GetUserByEmail :one
select id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at from users
where email = $1;

-- name: GetUsersByRole :many
select id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at from users
where role = $1;

-- name: BlockUserByID :one
update users
set is_blocked = true, updated_at = current_timestamp
where id = $1
returning id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at;

-- name: UnblockUserByID :one
update users
set is_blocked = false, updated_at = current_timestamp
where id = $1
returning id, name, email, phone, role, is_blocked, gst_no, about, created_at, updated_at;

-- name: GetOTPByUserID :one
select * from login_otps
where user_id = $1
order by created_at desc
limit 1;
