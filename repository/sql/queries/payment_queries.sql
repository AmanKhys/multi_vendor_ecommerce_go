-- name: AddPayment :one
insert into payments
(order_id, method, status, total_amount)
values
($1, $2, $3, $4)
returning *;

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

-- name: EditPaymentStatusByOrderID :one
update payments
set status = $2, updated_at = current_timestamp
where order_id = $1
returning *;


-- name: CancelPaymentByOrderID :exec
update payments
set status = 'cancelled', total_amount = 0, updated_at = current_timestamp
where order_id = $1;

-- name: AddVendorPayment :one
insert into vendor_payments
(order_item_id, seller_id, status, total_amount, platform_fee, credit_amount)
values
($1, $2, $3, $4, $5, $6)
returning *;

-- name: GetVendorPaymentByOrderItemID :one
select * from vendor_payments
where order_item_id = $1;