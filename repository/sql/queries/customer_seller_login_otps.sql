-- name: MakeAndGetSellerOTP :one
insert into seller_login_otps
(seller_id)
values ($1)
returning *;

-- name: MakeAndGetCustomerOTP :one
insert into customer_login_otps
(customer_id)
values ($1)
returning *;

