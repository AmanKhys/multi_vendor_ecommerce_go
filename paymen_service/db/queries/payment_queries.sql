-- name: AddPayment :one
insert into payments
(order_id, method, status, total_amount)
values
($1, $2, $3, $4)
returning *;

-- name: GetPaymentByOrderID :one
select * from payments
where order_id = @order_id;

-- name: DecPaymentAmountByOrderItemID :one
WITH cte AS (
  SELECT oi.order_id, oi.total_amount
  FROM order_items oi
  WHERE oi.id = $1
)
UPDATE payments
SET total_amount = payments.total_amount - cte.total_amount
FROM cte
WHERE payments.order_id = cte.order_id
RETURNING payments.*;


-- name: EditPaymentStatusByID :one
update payments
set status = $2, updated_at = current_timestamp
where id = $1
returning *;

-- name: EditPaymentByOrderID :one
update payments
set status = $2, transaction_id = $3, updated_at = current_timestamp
where order_id = $1
returning *;

-- name: EditPaymentStatusByOrderID :one
update payments
set status = $2, updated_at = current_timestamp
where order_id = $1
returning *;


-- name: CancelPaymentByOrderID :one
update payments
set status = 'cancelled', updated_at = current_timestamp
where order_id = $1
returning *;

-- name: AddVendorPayment :one
insert into vendor_payments
(order_item_id, seller_id, status, total_amount, platform_fee, credit_amount)
values
($1, $2, $3, $4, $5, $6)
returning *;

-- name: GetVendorPaymentByOrderItemID :one
select * from vendor_payments
where order_item_id = $1;

-- name: EditVendorPaymentStatusByOrderItemID :one
update vendor_payments
set status = $2
where order_item_id = $1
returning *;

-- name: GetVendorPaymentsBySellerID :many
select * from vendor_payments
where seller_id = $1;

-- name: GetVendorPaymentsBySellerIDAndDateRange :many
select * 
from vendor_payments
where seller_id = $1  and
created_at between @start_date and @end_date;

-- name: GetVendorPaymentsByDateRange :many
select * 
from vendor_payments
where 
created_at between @start_date and @end_date;

-- name: CancelVendorPaymentsByOrderID :many
update vendor_payments
set status = 'cancelled', updated_at = current_timestamp
where order_item_id in 
(select id from order_items where order_id = @order_id)
returning *;

-- name: CancelVendorPaymentByOrderItemID :exec
update vendor_payments
set status = 'cancelled', updated_at = current_timestamp
where order_item_id = $1;