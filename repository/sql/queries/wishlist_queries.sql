-- name: GetAllWishListItemsWithProductNameByUserID :many
select w.*, p.name as product_name from wishlists w
inner join products p
on w.product_id = p.id
where w.user_id = @user_id;

-- name: GetAllWishListItemsByUserID :many
select * from wishlists
where user_id = $1;

-- name: GetWishListItemByUserAndProductID :one
select * from wishlists
where user_id = $1 and product_id = $2;

-- name: DeleteWishListItemByUserAndProductID :execrows
delete from wishlists
where user_id = $1 and product_id = $2;

-- name: AddWishListItem :one
insert into wishlists
(user_id, product_id)
values
($1, $2)
returning *;

-- name: DeleteAllWishListItemsByUserID :exec
delete from wishlists
where user_id = $1;

-- name: AddAllWishListItemsToCarts :many
with cte AS
(select w.* from wishlists w where w.user_id = @user_id)
insert into carts
(user_id, product_id, quantity)
select user_id, product_id, 1 from cte
returning *;
