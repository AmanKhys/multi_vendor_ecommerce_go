-- name: AddWalletByUserID :one
insert into wallets
(user_id, savings)
values ($1, 0)
returning id, savings;

-- name: GetWalletByUserID :one
select id, savings from wallets
where user_id = $1;

-- name: AddSavingsToWalletByUserID :one
update wallets
set savings = savings + @savings, updated_at = current_timestamp
where user_id = @user_id and (savings + @savings) >= 0
returning id, savings;

-- name: RetractSavingsFromWalletByUserID :one
update wallets
set savings = savings - @savings, updated_at = current_timestamp
where user_id = @user_id
returning id, savings;