package admin

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var a = Admin{DB: DB}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/users/", a.AdminUserHandler)
	mux.HandleFunc("PUT /admin/user/{userID}/block/", a.BlockUserHandler)
	mux.HandleFunc("PUT /admin/user/{userID}/unblock/", a.UnblockUserHandler)

	mux.HandleFunc("GET /admin/sellers/", a.AdminSellerHandler)
	mux.HandleFunc("PUT /admin/seller/{sellerID}/block/", a.BlockSellerHandler)
	mux.HandleFunc("PUT /admin/seller/{sellerID}/unblock/", a.UnblockSellerHandler)

	mux.HandleFunc("GET /admin/products/", a.AdminProductsHandler)
	mux.HandleFunc("DELETE /admin/product/{productID}/delete/", a.DeleteProductHandler)

	mux.HandleFunc("GET /admin/categories/", a.AdminCategoriesHandler)
	mux.HandleFunc("POST /admin/category/add/", a.AddCategoryHandler)
	mux.HandleFunc("PUT /admin/category/edit/", a.EditCategoryHandler)
	mux.HandleFunc("DELETE /admin/category/delete/", a.DeleteCategoryHandler)
}
