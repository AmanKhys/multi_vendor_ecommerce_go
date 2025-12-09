-- name: GetAllCategoriesForAdmin :many
select * from categories;

-- name: GetAllCategories :many
select * from categories
where is_deleted = false;

-- name: GetCategoryByID :one
select * from categories
where id = $1 and is_deleted = false;

-- name: GetCategoryByName :one
select * from categories
where name = $1 and is_deleted = false;

-- name: AddCateogry :one
insert into categories
(name) values ($1)
returning *;

-- name: DeleteCategoryByName :one
update categories
set is_deleted = true, updated_at = current_timestamp
where name = $1
returning *;

-- name: EditCategoryNameByName :one
update categories
set name = @new_name, updated_at = current_timestamp
where name = @name and is_deleted = false
returning *;

-- name: DeleteAllCategoriesForProductByID :exec
delete from category_items
where product_id = $1;

-- name: GetCategoryNamesOfProductByID :many
select c.name from category_items ci
inner join categories c
on ci.category_id = c.id
where ci.product_id = $1 and c.is_deleted = false;

-- name: GetProductsByCategoryName :many
select p.id, p.name, p.description, p.price, p.stock, p.seller_id, p.created_at, p.updated_at from category_items ci
inner join products p
on ci.product_id = p.id
inner join categories c
on ci.category_id = c.id
where c.name = $1 and c.is_deleted = false and p.is_deleted = false;