-- name: AddAddressByUserID :one
insert into addresses
(user_id, type, building_name, street_name, town, district, state, pincode)
values
($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: GetAddressByID :one
select * from addresses
where id = $1;

-- name: GetAddressesByUserID :many
select * from addresses
where user_id = $1;

-- name: EditAddressByID :one
update addresses
set building_name = $2, street_name = $3, town = $4, district = $5, state = $6, pincode = $7, updated_at = current_timestamp
where id = $1
returning *;

-- name: DeleteAddressByID :exec
delete from addresses
where id = $1;

-- name: DeleteAddressesByUserID :exec
delete from addresses
where user_id = $1;