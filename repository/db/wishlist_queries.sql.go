// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: wishlist_queries.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const addAllWishListItemsToCarts = `-- name: AddAllWishListItemsToCarts :many
with cte AS
(select w.id, w.user_id, w.product_id, w.created_at from wishlists w where w.user_id = $1)
insert into carts
(user_id, product_id, quantity)
select user_id, product_id, 1 from cte
returning id, user_id, product_id, quantity, created_at, updated_at
`

func (q *Queries) AddAllWishListItemsToCarts(ctx context.Context, userID uuid.UUID) ([]Cart, error) {
	rows, err := q.query(ctx, q.addAllWishListItemsToCartsStmt, addAllWishListItemsToCarts, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Cart{}
	for rows.Next() {
		var i Cart
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ProductID,
			&i.Quantity,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const addWishListItem = `-- name: AddWishListItem :one
insert into wishlists
(user_id, product_id)
values
($1, $2)
returning id, user_id, product_id, created_at
`

type AddWishListItemParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
}

func (q *Queries) AddWishListItem(ctx context.Context, arg AddWishListItemParams) (Wishlist, error) {
	row := q.queryRow(ctx, q.addWishListItemStmt, addWishListItem, arg.UserID, arg.ProductID)
	var i Wishlist
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.CreatedAt,
	)
	return i, err
}

const deleteAllWishListItemsByUserID = `-- name: DeleteAllWishListItemsByUserID :exec
delete from wishlists
where user_id = $1
`

func (q *Queries) DeleteAllWishListItemsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := q.exec(ctx, q.deleteAllWishListItemsByUserIDStmt, deleteAllWishListItemsByUserID, userID)
	return err
}

const deleteWishListItemByUserAndProductID = `-- name: DeleteWishListItemByUserAndProductID :execrows
delete from wishlists
where user_id = $1 and product_id = $2
`

type DeleteWishListItemByUserAndProductIDParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
}

func (q *Queries) DeleteWishListItemByUserAndProductID(ctx context.Context, arg DeleteWishListItemByUserAndProductIDParams) (int64, error) {
	result, err := q.exec(ctx, q.deleteWishListItemByUserAndProductIDStmt, deleteWishListItemByUserAndProductID, arg.UserID, arg.ProductID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getAllWishListItemsByUserID = `-- name: GetAllWishListItemsByUserID :many
select id, user_id, product_id, created_at from wishlists
where user_id = $1
`

func (q *Queries) GetAllWishListItemsByUserID(ctx context.Context, userID uuid.UUID) ([]Wishlist, error) {
	rows, err := q.query(ctx, q.getAllWishListItemsByUserIDStmt, getAllWishListItemsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Wishlist{}
	for rows.Next() {
		var i Wishlist
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ProductID,
			&i.CreatedAt,
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

const getAllWishListItemsWithProductNameByUserID = `-- name: GetAllWishListItemsWithProductNameByUserID :many
select w.id, w.user_id, w.product_id, w.created_at, p.name as product_name from wishlists w
inner join products p
on w.product_id = p.id
where w.user_id = $1
`

type GetAllWishListItemsWithProductNameByUserIDRow struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ProductID   uuid.UUID `json:"product_id"`
	CreatedAt   time.Time `json:"created_at"`
	ProductName string    `json:"product_name"`
}

func (q *Queries) GetAllWishListItemsWithProductNameByUserID(ctx context.Context, userID uuid.UUID) ([]GetAllWishListItemsWithProductNameByUserIDRow, error) {
	rows, err := q.query(ctx, q.getAllWishListItemsWithProductNameByUserIDStmt, getAllWishListItemsWithProductNameByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllWishListItemsWithProductNameByUserIDRow{}
	for rows.Next() {
		var i GetAllWishListItemsWithProductNameByUserIDRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ProductID,
			&i.CreatedAt,
			&i.ProductName,
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

const getWishListItemByUserAndProductID = `-- name: GetWishListItemByUserAndProductID :one
select id, user_id, product_id, created_at from wishlists
where user_id = $1 and product_id = $2
`

type GetWishListItemByUserAndProductIDParams struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
}

func (q *Queries) GetWishListItemByUserAndProductID(ctx context.Context, arg GetWishListItemByUserAndProductIDParams) (Wishlist, error) {
	row := q.queryRow(ctx, q.getWishListItemByUserAndProductIDStmt, getWishListItemByUserAndProductID, arg.UserID, arg.ProductID)
	var i Wishlist
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ProductID,
		&i.CreatedAt,
	)
	return i, err
}
