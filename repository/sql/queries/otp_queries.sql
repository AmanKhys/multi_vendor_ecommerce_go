-- name: GetValidOTPByUserID :one
SELECT * FROM otps
WHERE user_id = $1 and expires_at > current_timestamp
ORDER BY created_at DESC
LIMIT 1;

-- name: AddOTP :one
insert into otps
(user_id) values ($1)
returning *;

-- name: DeleteOTPByEmail :execresult
delete from otps
where user_id = (select user_id from users where email = $1);

-- name: GetValidForgotOTPByUserID :one
select * from forgot_otps
where user_id  = $1 and expires_at > current_timestamp
order by created_at DESC
limit 1;

-- name: AddForgotOTPByUserID :one
insert into forgot_otps 
(user_id) values ($1)
returning *;

-- name: DeleteForgotOTPByEmail :exec
delete from forgot_otps
where user_id = (select user_id from users where email = $1);