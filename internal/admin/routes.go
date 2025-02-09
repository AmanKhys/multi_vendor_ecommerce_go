package admin

import (
	"net/http"

	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var a = Admin{DB: DB}

const AdminRole string = "admin"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/allusers", middleware.AuthenticateUserMiddleware(a.AdminAllUsersHandler, AdminRole))
	mux.HandleFunc("PUT /admin/user/block", middleware.AuthenticateUserMiddleware(a.BlockUserHandler, AdminRole))
	mux.HandleFunc("PUT /admin/user/unblock", middleware.AuthenticateUserMiddleware(a.UnblockUserHandler, AdminRole))

	mux.HandleFunc("GET /admin/users", middleware.AuthenticateUserMiddleware(a.AdminUsersHandler, AdminRole))
	mux.HandleFunc("GET /admin/sellers", middleware.AuthenticateUserMiddleware(a.AdminSellersHandler, AdminRole))
	mux.HandleFunc("POST /admin/verify_seller", middleware.AuthenticateUserMiddleware(a.VerifySellerHandler, AdminRole))

	mux.HandleFunc("GET /admin/products", middleware.AuthenticateUserMiddleware(a.AdminProductsHandler, AdminRole))
	mux.HandleFunc("DELETE /admin/product/delete", middleware.AuthenticateUserMiddleware(a.DeleteProductHandler, AdminRole))

	mux.HandleFunc("GET /admin/categories", middleware.AuthenticateUserMiddleware(a.AdminCategoriesHandler, AdminRole))
	mux.HandleFunc("POST /admin/category/add", middleware.AuthenticateUserMiddleware(a.AddCategoryHandler, AdminRole))
	mux.HandleFunc("PUT /admin/category/edit", middleware.AuthenticateUserMiddleware(a.EditCategoryHandler, AdminRole))
	mux.HandleFunc("DELETE /admin/category/delete", middleware.AuthenticateUserMiddleware(a.DeleteCategoryHandler, AdminRole))
}
