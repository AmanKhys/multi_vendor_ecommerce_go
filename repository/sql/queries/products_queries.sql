-- name: GetAllProductsForAdmin :many
select * from products;

-- name: AddProduct :one
insert into products
(name, description, price, stock, seller_id)
values ($1, $2, $3, $4, $5)
returning *;

-- name: GetProductByID :one
select * from products
where id = $1 and is_deleted = false;

-- name: GetAllProducts :many
select * from products
where is_deleted = false;

-- name: GetProductsBySellerID :many
select * from products
where seller_id = $1 and is_deleted = false;

-- name: EditProductByID :one
update products
set name = $2, description = $3, price = $4, stock = $5, updated_at = current_timestamp
where id = $1 and is_deleted = false
returning *;

-- name: DeleteProductByID :one
update products
set is_deleted = true, updated_at = current_timestamp
where id = $1 and is_deleted = false
returning *;

-- name: DeleteProductsBySellerID :many
update products
set is_deleted = true, updated_at = current_timestamp
where seller_id = $1
returning *;

-- name: AddProductToCategoryByID :one
insert into category_items
(product_id, category_id)
values
($1, $2)
returning *;

-- name: AddProductToCategoryByCategoryName :one
insert into category_items
(product_id, category_id)
values
(@product_id, (select id from categories where name = @category_name))
returning *;

-- name: GetProductAndCategoryNameByID :one
select p.*, c.name as category_name
from category_items ci
inner join products p
on ci.product_id = p.id
inner join categories c
on ci.category_id = c.id
where p.id = $1;