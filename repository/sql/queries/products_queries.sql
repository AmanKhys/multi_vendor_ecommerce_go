-- name: InsertProduct :one
insert into products
(name, description, price, stock, seller_id)
values ($1, $2, $3, $4, $5)
returning *;

-- name: GetAllProducts :many
select * from products;

-- name: GetProductsByCategoryID :many
select * from products
where category_id = $1;

-- name: GetProductsBySellerID :many
select * from products
where seller_id = $1;

-- name: UpdateProductByID :one
update products
set name = $2, description = $3, price = $4, stock = $5, updated_at = $6
where id = $1
returning *;

-- name: DeleteProductByID :one
update products
set is_deleted = true
where id = $1
returning *;

-- name: DeleteProductsBySellerID :many
update products
set is_deleted = true
where seller_id = $1
returning *;

