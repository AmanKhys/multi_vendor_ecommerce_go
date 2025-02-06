-- name: CreateNewSessionByUserID :one
insert into sessions
(user_id) 
values ($1)
returning *;

-- name: GetSessionDetailsByID :one
select * from sessions
where id = $1;

-- name: GetAllSessionsByUserID :one
select * from sessions
where user_id = $1;

-- name: GetUserBySessionID :one
select u.* from sessions s
join users u
on s.user_id = u.id
where u.id = $1;