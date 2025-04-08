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

-- name: EditCouponByID :one
update coupons
set name = @name, trigger_price = @trigger_price, discount_amount = @discount_amount,
updated_at = current_timestamp
returning *;

-- name: DeleteCouponByID :exec
delete from coupons
where id = $1;

-- name: AddCoupon :one
insert into coupons
(name, trigger_price, discount_amount)
values
($1, $2, $3)
returning *;

-- name: DeleteCouponByName :one
update coupons
set is_deleted = true, updated_at = current_timestamp
where name = $1
returning *;

-- name: EditCouponByName :one
update coupons
set name = @new_name, trigger_price = @trigger_price,
 discount_amount = @discount_amount, updated_at = current_timestamp
where name = @old_name
returning *;