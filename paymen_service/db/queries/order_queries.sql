-- name: GetAllOrders :many
select * from orders;

-- name: GetAllOrderItemsForAdmin :many
select * from order_items
order by created_at desc;

-- name: GetOrdersByUserID :many
select * from orders
where user_id = $1
order by created_at desc;

-- name: GetOrderByID :one
select * from orders
where id = $1;

-- name: GetOrderItemsBySellerID :many
select oi.* from order_items oi
inner join products p
on oi.product_id = p.id
where p.seller_id = $1
order by oi.created_at desc;

-- name: GetSellerIDFromOrderItemID :one
select p.seller_id from order_items oi
inner join products p
on oi.product_id = p.id
where oi.id = $1;


-- name: AddOrder :one
insert into orders
(user_id)
values (@user_id)
returning *;

-- name: UpdateOrderTotalAmount :one
update orders
set total_amount = @total_amount, updated_at = current_timestamp
where id = @id
returning *;

-- name: EditOrderAmountByID :one
update orders
set total_amount = @total_amount, discount_amount = @discount_amount, coupon_id = @coupon_id, updated_at = current_timestamp
where id = @id
returning *;

-- name: DeleteOrderByID :exec
delete from orders
where id = $1;

-- name: AddOrderITem :one
insert into order_items
(order_id, product_id, price, quantity)
values
($1, $2, $3, $4)
returning *;

-- name: GetOrderItemsByUserID :many
select oi.* from order_items oi
inner join orders o
on oi.order_id = o.id
where o.user_id = $1
order by oi.created_at desc;

-- name: GetOrderItemByID :one
select * from order_items
where id = $1;

-- name: GetUserIDFromOrderItemID :one
select o.user_id from order_items oi
inner join orders o
on oi.order_id = o.id
where oi.id = $1;

-- name: GetOrderItemsByOrderID :many
select oi.*, p.name as product_name
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

-- name: GetShippingAddressByOrderID :one
select id, house_name, street_name, town, district, state, pincode
from shipping_address
where order_id = $1;

-- name: ChangeOrderItemStatusByID :one
update order_items oi
set status =  $2
where id = $1
returning *;

-- name: EditOrderItemStatusByID :one
update order_items
set status = $2, updated_at = current_timestamp
where id = $1
returning *;

-- name: CancelOrderByID :many
update order_items
set status = 'cancelled', updated_at = current_timestamp
where order_id = $1
returning *;

-- name: GetTotalAmountOfCartItems :one
select sum(total_amount) as total_amount
from carts
where user_id = $1;

-- name: GetOrderItemsBySellerIDAndDateRange :many
select oi.* 
from order_items oi
inner join products p on oi.product_id = p.id
where p.seller_id = $1 
  and oi.created_at between @start_date and @end_date
order by oi.created_at desc;

-- name: GetOrderItemByUserAndProductID :one
select oi.*
from order_items oi
inner join orders o
on oi.order_id = o.id
where oi.product_id = @product_id and 
o.user_id = @user_id and
(oi.status = 'delivered' or oi.status = 'returned')
limit 1;

-- name: GetReviewByUserAndProductID :one
select *
from reviews r
where r.user_id = $1 and r.product_id = $2;