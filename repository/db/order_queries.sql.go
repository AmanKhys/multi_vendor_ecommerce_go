// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: order_queries.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const addOrder = `-- name: AddOrder :one
insert into orders
(user_id)
values ($1)
returning id, user_id, created_at, updated_at
`

func (q *Queries) AddOrder(ctx context.Context, userID uuid.UUID) (Order, error) {
	row := q.queryRow(ctx, q.addOrderStmt, addOrder, userID)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const addOrderITem = `-- name: AddOrderITem :one
insert into order_items
(order_id, product_id, quantity, total_amount)
values
($1, $2, $3, $4)
returning id, order_id, product_id, quantity, total_amount, status, created_at, updated_at
`

type AddOrderITemParams struct {
	OrderID     uuid.UUID `json:"order_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int32     `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
}

func (q *Queries) AddOrderITem(ctx context.Context, arg AddOrderITemParams) (OrderItem, error) {
	row := q.queryRow(ctx, q.addOrderITemStmt, addOrderITem,
		arg.OrderID,
		arg.ProductID,
		arg.Quantity,
		arg.TotalAmount,
	)
	var i OrderItem
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.ProductID,
		&i.Quantity,
		&i.TotalAmount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

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

const addShippingAddress = `-- name: AddShippingAddress :one
insert into shipping_address
(order_id, house_name, street_name, town, district, state, pincode)
values
($1, $2, $3, $4, $5,$6, $7)
returning id, house_name, street_name, town, district, state, pincode
`

type AddShippingAddressParams struct {
	OrderID    uuid.UUID `json:"order_id"`
	HouseName  string    `json:"house_name"`
	StreetName string    `json:"street_name"`
	Town       string    `json:"town"`
	District   string    `json:"district"`
	State      string    `json:"state"`
	Pincode    int32     `json:"pincode"`
}

type AddShippingAddressRow struct {
	ID         uuid.UUID `json:"id"`
	HouseName  string    `json:"house_name"`
	StreetName string    `json:"street_name"`
	Town       string    `json:"town"`
	District   string    `json:"district"`
	State      string    `json:"state"`
	Pincode    int32     `json:"pincode"`
}

func (q *Queries) AddShippingAddress(ctx context.Context, arg AddShippingAddressParams) (AddShippingAddressRow, error) {
	row := q.queryRow(ctx, q.addShippingAddressStmt, addShippingAddress,
		arg.OrderID,
		arg.HouseName,
		arg.StreetName,
		arg.Town,
		arg.District,
		arg.State,
		arg.Pincode,
	)
	var i AddShippingAddressRow
	err := row.Scan(
		&i.ID,
		&i.HouseName,
		&i.StreetName,
		&i.Town,
		&i.District,
		&i.State,
		&i.Pincode,
	)
	return i, err
}

const deleteOrderByID = `-- name: DeleteOrderByID :exec
delete from orders
where id = $1
`

func (q *Queries) DeleteOrderByID(ctx context.Context, id uuid.UUID) error {
	_, err := q.exec(ctx, q.deleteOrderByIDStmt, deleteOrderByID, id)
	return err
}

const editOrderItemStatusByID = `-- name: EditOrderItemStatusByID :one
update order_items
set status = $2, updated_at = current_timestamp
where id = $1
returning id, order_id, product_id, quantity, total_amount, status, created_at, updated_at
`

type EditOrderItemStatusByIDParams struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (q *Queries) EditOrderItemStatusByID(ctx context.Context, arg EditOrderItemStatusByIDParams) (OrderItem, error) {
	row := q.queryRow(ctx, q.editOrderItemStatusByIDStmt, editOrderItemStatusByID, arg.ID, arg.Status)
	var i OrderItem
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.ProductID,
		&i.Quantity,
		&i.TotalAmount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const editPaymentStatusByOrderID = `-- name: EditPaymentStatusByOrderID :one
update payments
set status = $2, updated_at = current_timestamp
where id = $1
returning id, order_id, method, status, total_amount, transaction_id, created_at, updated_at
`

type EditPaymentStatusByOrderIDParams struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (q *Queries) EditPaymentStatusByOrderID(ctx context.Context, arg EditPaymentStatusByOrderIDParams) (Payment, error) {
	row := q.queryRow(ctx, q.editPaymentStatusByOrderIDStmt, editPaymentStatusByOrderID, arg.ID, arg.Status)
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

const getOrderItemsByOrderID = `-- name: GetOrderItemsByOrderID :many
select p.name as product_name , p.price, oi.quantity, oi.total_amount, oi.status
from order_items oi
inner join products p
on oi.product_id = p.id
where oi.order_id = $1
`

type GetOrderItemsByOrderIDRow struct {
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int32   `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

func (q *Queries) GetOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]GetOrderItemsByOrderIDRow, error) {
	rows, err := q.query(ctx, q.getOrderItemsByOrderIDStmt, getOrderItemsByOrderID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetOrderItemsByOrderIDRow{}
	for rows.Next() {
		var i GetOrderItemsByOrderIDRow
		if err := rows.Scan(
			&i.ProductName,
			&i.Price,
			&i.Quantity,
			&i.TotalAmount,
			&i.Status,
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
