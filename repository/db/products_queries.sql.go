// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: products_queries.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const addProduct = `-- name: AddProduct :one
insert into products
(name, description, price, stock, seller_id)
values ($1, $2, $3, $4, $5)
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

type AddProductParams struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int32     `json:"stock"`
	SellerID    uuid.UUID `json:"seller_id"`
}

func (q *Queries) AddProduct(ctx context.Context, arg AddProductParams) (Product, error) {
	row := q.queryRow(ctx, q.addProductStmt, addProduct,
		arg.Name,
		arg.Description,
		arg.Price,
		arg.Stock,
		arg.SellerID,
	)
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

const addProductToCategoryByCategoryName = `-- name: AddProductToCategoryByCategoryName :one
insert into category_items
(product_id, category_id)
values
($1, (select id from categories where name = $2))
returning id, product_id, category_id, created_at, updated_at
`

type AddProductToCategoryByCategoryNameParams struct {
	ProductID    uuid.UUID `json:"product_id"`
	CategoryName string    `json:"category_name"`
}

func (q *Queries) AddProductToCategoryByCategoryName(ctx context.Context, arg AddProductToCategoryByCategoryNameParams) (CategoryItem, error) {
	row := q.queryRow(ctx, q.addProductToCategoryByCategoryNameStmt, addProductToCategoryByCategoryName, arg.ProductID, arg.CategoryName)
	var i CategoryItem
	err := row.Scan(
		&i.ID,
		&i.ProductID,
		&i.CategoryID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const addProductToCategoryByID = `-- name: AddProductToCategoryByID :one
insert into category_items
(product_id, category_id)
values
($1, $2)
returning id, product_id, category_id, created_at, updated_at
`

type AddProductToCategoryByIDParams struct {
	ProductID  uuid.UUID `json:"product_id"`
	CategoryID uuid.UUID `json:"category_id"`
}

func (q *Queries) AddProductToCategoryByID(ctx context.Context, arg AddProductToCategoryByIDParams) (CategoryItem, error) {
	row := q.queryRow(ctx, q.addProductToCategoryByIDStmt, addProductToCategoryByID, arg.ProductID, arg.CategoryID)
	var i CategoryItem
	err := row.Scan(
		&i.ID,
		&i.ProductID,
		&i.CategoryID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const decProductStockByID = `-- name: DecProductStockByID :one
update products
set stock = stock - $1, updated_at = current_timestamp
where id = $2 and stock >= $1
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

type DecProductStockByIDParams struct {
	DecQuantity int32     `json:"dec_quantity"`
	ProductID   uuid.UUID `json:"product_id"`
}

func (q *Queries) DecProductStockByID(ctx context.Context, arg DecProductStockByIDParams) (Product, error) {
	row := q.queryRow(ctx, q.decProductStockByIDStmt, decProductStockByID, arg.DecQuantity, arg.ProductID)
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

const deleteProductByID = `-- name: DeleteProductByID :one
update products
set is_deleted = true, updated_at = current_timestamp
where id = $1 and is_deleted = false
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

func (q *Queries) DeleteProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	row := q.queryRow(ctx, q.deleteProductByIDStmt, deleteProductByID, id)
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

const deleteProductsBySellerID = `-- name: DeleteProductsBySellerID :many
update products
set is_deleted = true, updated_at = current_timestamp
where seller_id = $1
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

func (q *Queries) DeleteProductsBySellerID(ctx context.Context, sellerID uuid.UUID) ([]Product, error) {
	rows, err := q.query(ctx, q.deleteProductsBySellerIDStmt, deleteProductsBySellerID, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.Stock,
			&i.SellerID,
			&i.IsDeleted,
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

const editProductByID = `-- name: EditProductByID :one
update products
set name = $2, description = $3, price = $4, stock = $5, updated_at = current_timestamp
where id = $1 and is_deleted = false
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

type EditProductByIDParams struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int32     `json:"stock"`
}

func (q *Queries) EditProductByID(ctx context.Context, arg EditProductByIDParams) (Product, error) {
	row := q.queryRow(ctx, q.editProductByIDStmt, editProductByID,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Price,
		arg.Stock,
	)
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

const getAllProducts = `-- name: GetAllProducts :many
select id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at from products
where is_deleted = false
`

func (q *Queries) GetAllProducts(ctx context.Context) ([]Product, error) {
	rows, err := q.query(ctx, q.getAllProductsStmt, getAllProducts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.Stock,
			&i.SellerID,
			&i.IsDeleted,
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

const getAllProductsForAdmin = `-- name: GetAllProductsForAdmin :many
select id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at from products
`

func (q *Queries) GetAllProductsForAdmin(ctx context.Context) ([]Product, error) {
	rows, err := q.query(ctx, q.getAllProductsForAdminStmt, getAllProductsForAdmin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.Stock,
			&i.SellerID,
			&i.IsDeleted,
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

const getProductAndCategoryNameByID = `-- name: GetProductAndCategoryNameByID :one
select p.id, p.name, p.description, p.price, p.stock, p.seller_id, p.is_deleted, p.created_at, p.updated_at, c.name as category_name
from category_items ci
inner join products p
on ci.product_id = p.id
inner join categories c
on ci.category_id = c.id
where p.id = $1
`

type GetProductAndCategoryNameByIDRow struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Stock        int32     `json:"stock"`
	SellerID     uuid.UUID `json:"seller_id"`
	IsDeleted    bool      `json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CategoryName string    `json:"category_name"`
}

func (q *Queries) GetProductAndCategoryNameByID(ctx context.Context, id uuid.UUID) (GetProductAndCategoryNameByIDRow, error) {
	row := q.queryRow(ctx, q.getProductAndCategoryNameByIDStmt, getProductAndCategoryNameByID, id)
	var i GetProductAndCategoryNameByIDRow
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
		&i.CategoryName,
	)
	return i, err
}

const getProductByID = `-- name: GetProductByID :one
select id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at from products
where id = $1 and is_deleted = false
`

func (q *Queries) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	row := q.queryRow(ctx, q.getProductByIDStmt, getProductByID, id)
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

const getProductsBySellerID = `-- name: GetProductsBySellerID :many
select id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at from products
where seller_id = $1 and is_deleted = false
`

func (q *Queries) GetProductsBySellerID(ctx context.Context, sellerID uuid.UUID) ([]Product, error) {
	rows, err := q.query(ctx, q.getProductsBySellerIDStmt, getProductsBySellerID, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.Stock,
			&i.SellerID,
			&i.IsDeleted,
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

const incProductStockByID = `-- name: IncProductStockByID :one
update products
set stock = stock + $1, updated_at = current_timestamp
where id = $2
returning id, name, description, price, stock, seller_id, is_deleted, created_at, updated_at
`

type IncProductStockByIDParams struct {
	IncQuantity int32     `json:"inc_quantity"`
	ProductID   uuid.UUID `json:"product_id"`
}

func (q *Queries) IncProductStockByID(ctx context.Context, arg IncProductStockByIDParams) (Product, error) {
	row := q.queryRow(ctx, q.incProductStockByIDStmt, incProductStockByID, arg.IncQuantity, arg.ProductID)
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
