-- name: AddOrder :one
insert into orders
(user_id)
values (@user_id)
returning *;

-- name: AddShippingAddress :one
insert into shipping_address
(order_id, house_name, street_name, town, district, state, pincode)
values
($1, $2, $3, $4, $5,$6, $7)
returning *;