-- name: GetAllProducts :many
select * from Products;

-- name: GetProductByID :one
select * from products
where id = $1;

-- name: AddProduct :one
insert into products
(seller_id, name, price, description, stock)
values ($1, $2, $3, $4, $5)
returning *;

-- name: EditProduct :one
update products
set name = $2, price = $3, description = $4, stock = $5
where id = $1
returning *;

-- name: SoftDeleteProductByID :one
update products
set is_deleted = true
where id = $1
returning *;

-- name: UndoDeletedProductByID :one
update products
set is_deleted = false
where id = $1
returning *;
