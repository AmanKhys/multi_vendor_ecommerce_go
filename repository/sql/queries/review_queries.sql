-- name: AddReview :one
insert into reviews
(customer_id, product_id, rating, comment)
values
($1, $2, $3, $4)
returning *;

-- name: GetProductReview :many
select * from reviews
where product_id = $1;

-- name: SoftDeleteReviewByID :one
update reviews
set is_deleted = true
where id = $1
returning *;

