// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: wallet_queries.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const addSavingsToWalletByUserID = `-- name: AddSavingsToWalletByUserID :one
update wallets
set savings = savings + $2, updated_at = current_timestamp
where user_id = $1 and (savings + $2) >= 0
returning id, savings
`

type AddSavingsToWalletByUserIDParams struct {
	UserID  uuid.UUID `json:"user_id"`
	Savings float64   `json:"savings"`
}

type AddSavingsToWalletByUserIDRow struct {
	ID      uuid.UUID `json:"id"`
	Savings float64   `json:"savings"`
}

func (q *Queries) AddSavingsToWalletByUserID(ctx context.Context, arg AddSavingsToWalletByUserIDParams) (AddSavingsToWalletByUserIDRow, error) {
	row := q.queryRow(ctx, q.addSavingsToWalletByUserIDStmt, addSavingsToWalletByUserID, arg.UserID, arg.Savings)
	var i AddSavingsToWalletByUserIDRow
	err := row.Scan(&i.ID, &i.Savings)
	return i, err
}

const addWalletByUserID = `-- name: AddWalletByUserID :one
insert into wallets
(user_id, savings)
values ($1, 0)
returning id, savings
`

type AddWalletByUserIDRow struct {
	ID      uuid.UUID `json:"id"`
	Savings float64   `json:"savings"`
}

func (q *Queries) AddWalletByUserID(ctx context.Context, userID uuid.UUID) (AddWalletByUserIDRow, error) {
	row := q.queryRow(ctx, q.addWalletByUserIDStmt, addWalletByUserID, userID)
	var i AddWalletByUserIDRow
	err := row.Scan(&i.ID, &i.Savings)
	return i, err
}

const getWalletByUserID = `-- name: GetWalletByUserID :one
select id, savings from wallets
where user_id = $1
`

type GetWalletByUserIDRow struct {
	ID      uuid.UUID `json:"id"`
	Savings float64   `json:"savings"`
}

func (q *Queries) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (GetWalletByUserIDRow, error) {
	row := q.queryRow(ctx, q.getWalletByUserIDStmt, getWalletByUserID, userID)
	var i GetWalletByUserIDRow
	err := row.Scan(&i.ID, &i.Savings)
	return i, err
}

const retractSavingsFromWalletByUserID = `-- name: RetractSavingsFromWalletByUserID :one
update wallets
set savings = savings - $2, updated_at = current_timestamp
where user_id = $1
returning id, savings
`

type RetractSavingsFromWalletByUserIDParams struct {
	UserID  uuid.UUID `json:"user_id"`
	Savings float64   `json:"savings"`
}

type RetractSavingsFromWalletByUserIDRow struct {
	ID      uuid.UUID `json:"id"`
	Savings float64   `json:"savings"`
}

func (q *Queries) RetractSavingsFromWalletByUserID(ctx context.Context, arg RetractSavingsFromWalletByUserIDParams) (RetractSavingsFromWalletByUserIDRow, error) {
	row := q.queryRow(ctx, q.retractSavingsFromWalletByUserIDStmt, retractSavingsFromWalletByUserID, arg.UserID, arg.Savings)
	var i RetractSavingsFromWalletByUserIDRow
	err := row.Scan(&i.ID, &i.Savings)
	return i, err
}
