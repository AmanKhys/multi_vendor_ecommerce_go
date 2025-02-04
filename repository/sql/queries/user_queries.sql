-- name: InsertUser :one
insert into users
(name, email, phone, password, role, gst_no, about)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;
-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :many
select * from users
where id = $1;

-- name: GetUserByEmail :one
select * from users
where email = $1;

-- name: GetUsersByRole :many
select * from users
where role = $1;

-- name: BlockUserByID :one
update users
set is_blocked = true
where id = $1
returning *;

-- name: UnblockUserByID :one
update users
set is_blocked = false
where id = $1
returning *;

-- name: GetOTPByUserID :one
select * from login_otps
where user_id = $1
order by created_at desc
limit 1;
