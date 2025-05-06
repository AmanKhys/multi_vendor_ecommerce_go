-- name: GetCouponByID :one
select * from coupons
where id = $1;

-- name: GetAllCouponsForAdmin :many
select * from coupons;

-- name: GetAllCoupons :many
select * from coupons where is_deleted = false;

-- name: GetCouponByName :one
select * from coupons
where name = $1;

-- name: GetValidCouponByName :one
select * from coupons
where current_timestamp >= start_date and current_timestamp <= end_date
and name = @name;

-- name: EditCouponByID :one
update coupons
set name = @name, discount_type = @discount_type, trigger_price = @trigger_price, discount_amount = @discount_amount,
start_date = @start_date, end_date = @end_date
returning *;

-- name: DeleteCouponByID :exec
delete from coupons
where id = $1;

-- name: AddCoupon :one
insert into coupons
(name, discount_type, trigger_price, discount_amount, start_date, end_date)
values
($1, $2, $3, $4, $5, $6)
returning *;

-- name: DeleteCouponByName :one
update coupons
set is_deleted = true
where name = $1
returning *;

-- name: EditCouponByName :one
update coupons
set name = @new_name, trigger_price = @trigger_price,
 discount_type = @discount_type,
 discount_amount = @discount_amount,
 start_date = @start_date, end_date = @end_date
where name = @old_name
returning *;