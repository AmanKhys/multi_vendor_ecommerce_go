-- name: AddSeller :one
insert into sellers
(name, email, phone, password, about)
values
($1, $2, $3, $4, $5)
returning id, email, phone, password, is_blocked;

-- name: GetSellers :many
select s.id, s.name, s.email, s.phone, s.about, s.is_blocked, sad.* from sellers s
left join seller_addresses sad
on sad.seller_id = s.id;

-- name: GetSellerByID :one
select s.id, s.name, s.email, s.phone, s.about, s.is_blocked, sad.* from sellers s
left join seller_addresses sad
on s.id = sad.seller_id
where s.id = $1;

-- name: BlockSeller :one
update sellers
set is_blocked = true
where id = $1
returning id, name, email, phone, about, is_blocked;

-- name: UnblockSeller :one
update sellers
set is_blocked = false
where id = $1
returning id, name, email, phone, about, is_blocked;

-- name: UpdateSellerByID :one
update sellers
set name = $2, email = $3, phone = $4, password = $5, about = $6
where id = $1
returning id, name, email, phone, about, is_blocked;

