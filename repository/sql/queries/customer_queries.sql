-- name: AddCustomer :exec
insert into customers
(name, email, phone, password)
values
($1, $2, $3, $4);

-- name: GetAllCustomers :many
select c.name, c.email, c.phone, ad.* from customers c
left join customer_addresses ad
on ad.customer_id = c.id;

-- name: GetCustomerByID :one
select name, email, phone, ad.* from customers c
left join customer_addresses ad
on c.id = ad.id
where c.id = $1;

-- name: BlockCustomer :exec
update customers
set is_blocked = true
where id = $1;

-- name: UnblockCustomer :exec
update customers
set is_blocked = false
where id = $1;

-- name: UpdateCustomerByID :exec
update customers
set name = $2, email = $3, phone = $4, password = $5
where id = $1;

