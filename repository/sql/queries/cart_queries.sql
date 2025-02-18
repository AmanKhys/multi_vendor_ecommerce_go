-- name: GetCartItemsByUserID :many
select p.name as product_name, c.quantity from carts c
inner join products p
on c.product_id = p.id
where user_id = $1;

-- name: GetCartItemByUserIDAndProductID :one
select * from carts
where user_id = $1 and product_id = $2;

-- name: GetProductNameAndQuantityFromCartsByID :one
select p.name as product_name, c.quantity
from carts c
inner join products p
on c.product_id = p.id
where c.id = $1;

-- name: AddCartItem :one
insert into carts
(user_id, product_id, quantity)
values
($1, $2, $3)
returning *;

-- name: EditCartItemByID :one
update carts
set quantity = $2, updated_at = current_timestamp
where id = $1
returning *;

-- name: DeleteCartItemByUserIDAndProductID :exec
delete from carts
where user_id = $1 and product_id = $2;
