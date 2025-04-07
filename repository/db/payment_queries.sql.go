// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: payment_queries.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const addPayment = `-- name: AddPayment :one
insert into payments
(order_id, method, status, total_amount)
values
($1, $2, $3, $4)
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

type AddPaymentParams struct {
	OrderID     uuid.UUID `json:"order_id"`
	Method      string    `json:"method"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
}

func (q *Queries) AddPayment(ctx context.Context, arg AddPaymentParams) (Payment, error) {
	row := q.queryRow(ctx, q.addPaymentStmt, addPayment,
		arg.OrderID,
		arg.Method,
		arg.Status,
		arg.TotalAmount,
	)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const addVendorPayment = `-- name: AddVendorPayment :one
insert into vendor_payments
(order_item_id, seller_id, status, total_amount, platform_fee, credit_amount)
values
($1, $2, $3, $4, $5, $6)
returning id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at
`

type AddVendorPaymentParams struct {
	OrderItemID  uuid.UUID `json:"order_item_id"`
	SellerID     uuid.UUID `json:"seller_id"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	PlatformFee  float64   `json:"platform_fee"`
	CreditAmount float64   `json:"credit_amount"`
}

func (q *Queries) AddVendorPayment(ctx context.Context, arg AddVendorPaymentParams) (VendorPayment, error) {
	row := q.queryRow(ctx, q.addVendorPaymentStmt, addVendorPayment,
		arg.OrderItemID,
		arg.SellerID,
		arg.Status,
		arg.TotalAmount,
		arg.PlatformFee,
		arg.CreditAmount,
	)
	var i VendorPayment
	err := row.Scan(
		&i.ID,
		&i.OrderItemID,
		&i.SellerID,
		&i.Status,
		&i.TotalAmount,
		&i.PlatformFee,
		&i.CreditAmount,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const cancelPaymentByOrderID = `-- name: CancelPaymentByOrderID :one
update payments
set status = 'cancelled', updated_at = current_timestamp
where order_id = $1
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

func (q *Queries) CancelPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (Payment, error) {
	row := q.queryRow(ctx, q.cancelPaymentByOrderIDStmt, cancelPaymentByOrderID, orderID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const cancelVendorPaymentByOrderItemID = `-- name: CancelVendorPaymentByOrderItemID :exec
update vendor_payments
set status = 'cancelled', updated_at = current_timestamp
where order_item_id = $1
`

func (q *Queries) CancelVendorPaymentByOrderItemID(ctx context.Context, orderItemID uuid.UUID) error {
	_, err := q.exec(ctx, q.cancelVendorPaymentByOrderItemIDStmt, cancelVendorPaymentByOrderItemID, orderItemID)
	return err
}

const cancelVendorPaymentsByOrderID = `-- name: CancelVendorPaymentsByOrderID :many
update vendor_payments
set status = 'cancelled', updated_at = current_timestamp
where order_item_id in 
(select id from order_items where order_id = $1)
returning id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at
`

func (q *Queries) CancelVendorPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]VendorPayment, error) {
	rows, err := q.query(ctx, q.cancelVendorPaymentsByOrderIDStmt, cancelVendorPaymentsByOrderID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []VendorPayment{}
	for rows.Next() {
		var i VendorPayment
		if err := rows.Scan(
			&i.ID,
			&i.OrderItemID,
			&i.SellerID,
			&i.Status,
			&i.TotalAmount,
			&i.PlatformFee,
			&i.CreditAmount,
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

const decPaymentAmountByOrderItemID = `-- name: DecPaymentAmountByOrderItemID :one
WITH cte AS (
  SELECT oi.order_id, oi.total_amount
  FROM order_items oi
  WHERE oi.id = $1
)
UPDATE payments
SET total_amount = payments.total_amount - cte.total_amount
FROM cte
WHERE payments.order_id = cte.order_id
RETURNING payments.id, payments.order_id, payments.method, payments.status, payments.total_amount, payments.transaction_id, payments.created_at, payments.updated_at
`

func (q *Queries) DecPaymentAmountByOrderItemID(ctx context.Context, id uuid.UUID) (Payment, error) {
	row := q.queryRow(ctx, q.decPaymentAmountByOrderItemIDStmt, decPaymentAmountByOrderItemID, id)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const editPaymentByOrderID = `-- name: EditPaymentByOrderID :one
update payments
set status = $2, transaction_id = $3, updated_at = current_timestamp
where order_id = $1
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

type EditPaymentByOrderIDParams struct {
	OrderID       uuid.UUID      `json:"order_id"`
	Status        string         `json:"status"`
	TransactionID sql.NullString `json:"transaction_id"`
}

func (q *Queries) EditPaymentByOrderID(ctx context.Context, arg EditPaymentByOrderIDParams) (Payment, error) {
	row := q.queryRow(ctx, q.editPaymentByOrderIDStmt, editPaymentByOrderID, arg.OrderID, arg.Status, arg.TransactionID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const editPaymentStatusByID = `-- name: EditPaymentStatusByID :one
update payments
set status = $2, updated_at = current_timestamp
where id = $1
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

type EditPaymentStatusByIDParams struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (q *Queries) EditPaymentStatusByID(ctx context.Context, arg EditPaymentStatusByIDParams) (Payment, error) {
	row := q.queryRow(ctx, q.editPaymentStatusByIDStmt, editPaymentStatusByID, arg.ID, arg.Status)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const editPaymentStatusByOrderID = `-- name: EditPaymentStatusByOrderID :one
update payments
set status = $2, updated_at = current_timestamp
where order_id = $1
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

type EditPaymentStatusByOrderIDParams struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
}

func (q *Queries) EditPaymentStatusByOrderID(ctx context.Context, arg EditPaymentStatusByOrderIDParams) (Payment, error) {
	row := q.queryRow(ctx, q.editPaymentStatusByOrderIDStmt, editPaymentStatusByOrderID, arg.OrderID, arg.Status)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const editVendorPaymentStatusByOrderItemID = `-- name: EditVendorPaymentStatusByOrderItemID :one
update vendor_payments
set status = $2
where order_item_id = $1
returning id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at
`

type EditVendorPaymentStatusByOrderItemIDParams struct {
	OrderItemID uuid.UUID `json:"order_item_id"`
	Status      string    `json:"status"`
}

func (q *Queries) EditVendorPaymentStatusByOrderItemID(ctx context.Context, arg EditVendorPaymentStatusByOrderItemIDParams) (VendorPayment, error) {
	row := q.queryRow(ctx, q.editVendorPaymentStatusByOrderItemIDStmt, editVendorPaymentStatusByOrderItemID, arg.OrderItemID, arg.Status)
	var i VendorPayment
	err := row.Scan(
		&i.ID,
		&i.OrderItemID,
		&i.SellerID,
		&i.Status,
		&i.TotalAmount,
		&i.PlatformFee,
		&i.CreditAmount,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPaymentByOrderID = `-- name: GetPaymentByOrderID :one
select id, order_id, method, status, total_amount, transaction_id, created_at, updated_at from payments
where order_id = $1
`

func (q *Queries) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (Payment, error) {
	row := q.queryRow(ctx, q.getPaymentByOrderIDStmt, getPaymentByOrderID, orderID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Method,
		&i.Status,
		&i.TotalAmount,
		&i.TransactionID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getVendorPaymentByOrderItemID = `-- name: GetVendorPaymentByOrderItemID :one
select id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at from vendor_payments
where order_item_id = $1
`

func (q *Queries) GetVendorPaymentByOrderItemID(ctx context.Context, orderItemID uuid.UUID) (VendorPayment, error) {
	row := q.queryRow(ctx, q.getVendorPaymentByOrderItemIDStmt, getVendorPaymentByOrderItemID, orderItemID)
	var i VendorPayment
	err := row.Scan(
		&i.ID,
		&i.OrderItemID,
		&i.SellerID,
		&i.Status,
		&i.TotalAmount,
		&i.PlatformFee,
		&i.CreditAmount,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getVendorPaymentsByDateRange = `-- name: GetVendorPaymentsByDateRange :many
select id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at 
from vendor_payments
where 
created_at between $1 and $2
`

type GetVendorPaymentsByDateRangeParams struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func (q *Queries) GetVendorPaymentsByDateRange(ctx context.Context, arg GetVendorPaymentsByDateRangeParams) ([]VendorPayment, error) {
	rows, err := q.query(ctx, q.getVendorPaymentsByDateRangeStmt, getVendorPaymentsByDateRange, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []VendorPayment{}
	for rows.Next() {
		var i VendorPayment
		if err := rows.Scan(
			&i.ID,
			&i.OrderItemID,
			&i.SellerID,
			&i.Status,
			&i.TotalAmount,
			&i.PlatformFee,
			&i.CreditAmount,
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

const getVendorPaymentsBySellerID = `-- name: GetVendorPaymentsBySellerID :many
select id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at from vendor_payments
where seller_id = $1
`

func (q *Queries) GetVendorPaymentsBySellerID(ctx context.Context, sellerID uuid.UUID) ([]VendorPayment, error) {
	rows, err := q.query(ctx, q.getVendorPaymentsBySellerIDStmt, getVendorPaymentsBySellerID, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []VendorPayment{}
	for rows.Next() {
		var i VendorPayment
		if err := rows.Scan(
			&i.ID,
			&i.OrderItemID,
			&i.SellerID,
			&i.Status,
			&i.TotalAmount,
			&i.PlatformFee,
			&i.CreditAmount,
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

const getVendorPaymentsBySellerIDAndDateRange = `-- name: GetVendorPaymentsBySellerIDAndDateRange :many
select id, order_item_id, seller_id, status, total_amount, platform_fee, credit_amount, created_at, updated_at 
from vendor_payments
where seller_id = $1  and
created_at between $2 and $3
`

type GetVendorPaymentsBySellerIDAndDateRangeParams struct {
	SellerID  uuid.UUID `json:"seller_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func (q *Queries) GetVendorPaymentsBySellerIDAndDateRange(ctx context.Context, arg GetVendorPaymentsBySellerIDAndDateRangeParams) ([]VendorPayment, error) {
	rows, err := q.query(ctx, q.getVendorPaymentsBySellerIDAndDateRangeStmt, getVendorPaymentsBySellerIDAndDateRange, arg.SellerID, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []VendorPayment{}
	for rows.Next() {
		var i VendorPayment
		if err := rows.Scan(
			&i.ID,
			&i.OrderItemID,
			&i.SellerID,
			&i.Status,
			&i.TotalAmount,
			&i.PlatformFee,
			&i.CreditAmount,
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
