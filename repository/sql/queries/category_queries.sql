-- name: GetAllCategoriesForAdmin :many
select * from categories;

-- name: GetAllCategories :many
select * from categories
where is_deleted = false;

-- name: GetCategoryByID :one
select * from categories
where id = $1 and is_deleted = false;

-- name: AddCateogry :one
insert into categories
(name) values ($1)
returning *;

-- name: DeleteCategoryByID :one
update categories
set is_deleted = true, updated_at = current_timestamp
where id = $1
returning *;

-- name: EditCategoryNameByID :one
update categories
set name = $2, updated_at = current_timestamp
where id = $1
returning *;
