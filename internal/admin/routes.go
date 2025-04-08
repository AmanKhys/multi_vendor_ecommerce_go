package admin

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/helpers"
	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var a = Admin{DB: DB}
var helper = helpers.Helper{
	DB: DB,
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/allusers", middleware.AuthenticateUserMiddleware(a.AdminAllUsersHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/user/block", middleware.AuthenticateUserMiddleware(a.BlockUserHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/user/unblock", middleware.AuthenticateUserMiddleware(a.UnblockUserHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/users", middleware.AuthenticateUserMiddleware(a.AdminUsersHandler, utils.AdminRole))
	mux.HandleFunc("GET /admin/sellers", middleware.AuthenticateUserMiddleware(a.AdminSellersHandler, utils.AdminRole))
	mux.HandleFunc("POST /admin/verify_seller", middleware.AuthenticateUserMiddleware(a.VerifySellerHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/products", middleware.AuthenticateUserMiddleware(a.AdminProductsHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/product/delete", middleware.AuthenticateUserMiddleware(a.DeleteProductHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/categories", middleware.AuthenticateUserMiddleware(a.AdminCategoriesHandler, utils.AdminRole))
	mux.HandleFunc("POST /admin/category/add", middleware.AuthenticateUserMiddleware(a.AddCategoryHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/category/edit", middleware.AuthenticateUserMiddleware(a.EditCategoryHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/category/delete", middleware.AuthenticateUserMiddleware(a.DeleteCategoryHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/orders", middleware.AuthenticateUserMiddleware(a.GetOrderItemsHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/orders/deliver", middleware.AuthenticateUserMiddleware(a.DeliverOrderItemHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/coupons", middleware.AuthenticateUserMiddleware(a.AdminCouponsHandler, utils.AdminRole))
	mux.HandleFunc("POST /admin/coupons/add", middleware.AuthenticateUserMiddleware(a.AddCouponHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/coupons/edit", middleware.AuthenticateUserMiddleware(a.EditCouponHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/coupons/delete", middleware.AuthenticateUserMiddleware(a.DeleteCouponHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/sales_report", middleware.AuthenticateUserMiddleware(a.SalesReportHandler, utils.AdminRole))
}
