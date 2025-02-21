-- name: AddOrder :one
insert into orders
(user_id)
values (@user_id)
returning *;

-- name: DeleteOrderByID :exec
delete from orders
where id = $1;

-- name: AddOrderITem :one
insert into order_items
(order_id, product_id, quantity, total_amount)
values
($1, $2, $3, $4)
returning *;

-- name: GetOrderItemsByOrderID :many
select p.name as product_name , p.price, oi.quantity, oi.total_amount, oi.status
from order_items oi
inner join products p
on oi.product_id = p.id
where oi.order_id = $1;

-- name: AddShippingAddress :one
insert into shipping_address
(order_id, house_name, street_name, town, district, state, pincode)
values
($1, $2, $3, $4, $5,$6, $7)
returning id, house_name, street_name, town, district, state, pincode;

-- name: AddPayment :one
insert into payments
(order_id, method, status, total_amount)
values
($1, $2, $3, $4)
returning *;

-- name: EditOrderItemStatusByID :one
update order_items
set status = $2, updated_at = current_timestamp
where id = $1
returning *;

-- name: EditPaymentStatusByOrderID :one
update payments
set status = $2, updated_at = current_timestamp
where id = $1
returning *;