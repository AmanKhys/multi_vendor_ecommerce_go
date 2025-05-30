// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: cart_queries.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const addCartItem = `-- name: AddCartItem :one
insert into carts
(user_id, product_id, quantity)
values
($1, $2, $3)
returning id, user_id, product_id, quantity, created_at, updated_at
`

type AddCartItemParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

func (q *Queries) AddCartItem(ctx context.Context, arg AddCartItemParams) (Cart, error) {
	row := q.queryRow(ctx, q.addCartItemStmt, addCartItem, arg.UserID, arg.ProductID, arg.Quantity)
	var i Cart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.Quantity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteCartItemByUserIDAndProductID = `-- name: DeleteCartItemByUserIDAndProductID :exec
delete from carts
where user_id = $1 and product_id = $2
`

type DeleteCartItemByUserIDAndProductIDParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
}

func (q *Queries) DeleteCartItemByUserIDAndProductID(ctx context.Context, arg DeleteCartItemByUserIDAndProductIDParams) error {
	_, err := q.exec(ctx, q.deleteCartItemByUserIDAndProductIDStmt, deleteCartItemByUserIDAndProductID, arg.UserID, arg.ProductID)
	return err
}

const deleteCartItemsByUserID = `-- name: DeleteCartItemsByUserID :exec
delete from carts
where user_id = $1
`

func (q *Queries) DeleteCartItemsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := q.exec(ctx, q.deleteCartItemsByUserIDStmt, deleteCartItemsByUserID, userID)
	return err
}

const editCartItemByID = `-- name: EditCartItemByID :one
update carts
set quantity = $2, updated_at = current_timestamp
where id = $1
returning id, user_id, product_id, quantity, created_at, updated_at
`

type EditCartItemByIDParams struct {
	ID       uuid.UUID `json:"id"`
	Quantity int32     `json:"quantity"`
}

func (q *Queries) EditCartItemByID(ctx context.Context, arg EditCartItemByIDParams) (Cart, error) {
	row := q.queryRow(ctx, q.editCartItemByIDStmt, editCartItemByID, arg.ID, arg.Quantity)
	var i Cart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.Quantity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCartItemByID = `-- name: GetCartItemByID :one
select id, user_id, product_id, quantity, created_at, updated_at from carts
where id = $1
`

func (q *Queries) GetCartItemByID(ctx context.Context, id uuid.UUID) (Cart, error) {
	row := q.queryRow(ctx, q.getCartItemByIDStmt, getCartItemByID, id)
	var i Cart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.Quantity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCartItemByUserIDAndProductID = `-- name: GetCartItemByUserIDAndProductID :one
select id, user_id, product_id, quantity, created_at, updated_at from carts
where user_id = $1 and product_id = $2
`

type GetCartItemByUserIDAndProductIDParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
}

func (q *Queries) GetCartItemByUserIDAndProductID(ctx context.Context, arg GetCartItemByUserIDAndProductIDParams) (Cart, error) {
	row := q.queryRow(ctx, q.getCartItemByUserIDAndProductIDStmt, getCartItemByUserIDAndProductID, arg.UserID, arg.ProductID)
	var i Cart
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.Quantity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCartItemsByUserID = `-- name: GetCartItemsByUserID :many
select c.id as cart_id, p.id as product_id, p.name as product_name, c.quantity, p.price, (p.price * c.quantity)::numeric(10,2) as total_amount
from carts c
inner join products p
on c.product_id = p.id
where user_id = $1
`

type GetCartItemsByUserIDRow struct {
	CartID      uuid.UUID `json:"cart_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int32     `json:"quantity"`
	Price       float64   `json:"price"`
	TotalAmount float64   `json:"total_amount"`
}

func (q *Queries) GetCartItemsByUserID(ctx context.Context, userID uuid.UUID) ([]GetCartItemsByUserIDRow, error) {
	rows, err := q.query(ctx, q.getCartItemsByUserIDStmt, getCartItemsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetCartItemsByUserIDRow{}
	for rows.Next() {
		var i GetCartItemsByUserIDRow
		if err := rows.Scan(
			&i.CartID,
			&i.ProductID,
			&i.ProductName,
			&i.Quantity,
			&i.Price,
			&i.TotalAmount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProductFromCartByID = `-- name: GetProductFromCartByID :one
select p.id, p.name, p.description, p.price, p.stock, p.seller_id, p.is_deleted, p.created_at, p.updated_at from carts c
inner join products p
on c.product_id = p.id
where c.id = $1
`

func (q *Queries) GetProductFromCartByID(ctx context.Context, id uuid.UUID) (Product, error) {
	row := q.queryRow(ctx, q.getProductFromCartByIDStmt, getProductFromCartByID, id)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Price,
		&i.Stock,
		&i.SellerID,
		&i.IsDeleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getProductNameAndQuantityFromCartsByID = `-- name: GetProductNameAndQuantityFromCartsByID :one
select p.name as product_name, c.quantity
from carts c
inner join products p
on c.product_id = p.id
where c.id = $1
`

type GetProductNameAndQuantityFromCartsByIDRow struct {
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
}

func (q *Queries) GetProductNameAndQuantityFromCartsByID(ctx context.Context, id uuid.UUID) (GetProductNameAndQuantityFromCartsByIDRow, error) {
	row := q.queryRow(ctx, q.getProductNameAndQuantityFromCartsByIDStmt, getProductNameAndQuantityFromCartsByID, id)
	var i GetProductNameAndQuantityFromCartsByIDRow
	err := row.Scan(&i.ProductName, &i.Quantity)
	return i, err
}

const getSumOfCartItemsByUserID = `-- name: GetSumOfCartItemsByUserID :one
select cast(sum(p.price * cast(c.quantity as float)) as double precision) as total_amount
from carts c 
inner join products p on p.id = c.product_id
where c.user_id = $1
`

func (q *Queries) GetSumOfCartItemsByUserID(ctx context.Context, userID uuid.UUID) (float64, error) {
	row := q.queryRow(ctx, q.getSumOfCartItemsByUserIDStmt, getSumOfCartItemsByUserID, userID)
	var total_amount float64
	err := row.Scan(&total_amount)
	return total_amount, err
}
