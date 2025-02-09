-- name: GetSessionDetailsByID :one
select * from sessions
where id = $1;

-- name: GetAllSessionsByUserID :one
select * from sessions
where user_id = $1;

-- name: GetUserBySessionID :one
select 
    u.id, 
    u.name, 
    u.email, 
    u.phone, 
    u.role, 
    u.is_blocked, 
    u.gst_no, 
    u.about, 
    u.created_at, 
    u.updated_at
from sessions s
join users u
on s.user_id = u.id
where s.id = $1;


-- name: AddSession :one
insert into sessions
(user_id, ip_address, user_agent )
values
($1, $2, $3)
returning *;