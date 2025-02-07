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
	mux.HandleFunc("GET /admin/allusers", a.AdminAllUsersHandler)
	mux.HandleFunc("PUT /admin/user/block", a.BlockUserHandler)
	mux.HandleFunc("PUT /admin/user/unblock", a.UnblockUserHandler)

	mux.HandleFunc("GET /admin/users", a.AdminUsersHandler)
	mux.HandleFunc("GET /admin/sellers", a.AdminSellersHandler)

	mux.HandleFunc("GET /admin/products", a.AdminProductsHandler)
	mux.HandleFunc("DELETE /admin/product/delete", a.DeleteProductHandler)

	mux.HandleFunc("GET /admin/categories", a.AdminCategoriesHandler)
	mux.HandleFunc("POST /admin/category/add", a.AddCategoryHandler)
	mux.HandleFunc("PUT /admin/category/edit", a.EditCategoryHandler)
	mux.HandleFunc("DELETE /admin/category/delete", a.DeleteCategoryHandler)
}
