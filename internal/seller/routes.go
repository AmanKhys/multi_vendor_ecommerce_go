package seller

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
var s = Seller{DB: DB}
var helper = helpers.Helper{
	DB: DB,
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PUT /seller/profile/edit", middleware.AuthenticateUserMiddleware(s.EditProfileHandler, utils.SellerRole))

	mux.HandleFunc("GET /seller/products", middleware.AuthenticateUserMiddleware(s.OwnProductsHandler, utils.SellerRole))
	mux.HandleFunc("GET /seller/product", middleware.AuthenticateUserMiddleware(s.ProductDetailsHandler, utils.SellerRole))
	mux.HandleFunc("POST /seller/product/add", middleware.AuthenticateUserMiddleware(s.AddProductHandler, utils.SellerRole))
	mux.HandleFunc("PUT /seller/product/edit", middleware.AuthenticateUserMiddleware(s.EditProductHandler, utils.SellerRole))
	mux.HandleFunc("DELETE /seller/product/delete", middleware.AuthenticateUserMiddleware(s.DeleteProductHandler, utils.SellerRole))

	mux.HandleFunc("GET /seller/categories", middleware.AuthenticateUserMiddleware(s.GetAllCategoriesHandler, utils.SellerRole))
	mux.HandleFunc("POST /seller/category/add", middleware.AuthenticateUserMiddleware(s.AddProductToCategoryHandler, utils.SellerRole))

	mux.HandleFunc("GET /seller/address", middleware.AuthenticateUserMiddleware(s.GetAddressesHandler, utils.SellerRole))
	mux.HandleFunc("POST /seller/address/add", middleware.AuthenticateUserMiddleware(s.AddAddressHandler, utils.SellerRole))
	mux.HandleFunc("PUT /seller/address/edit", middleware.AuthenticateUserMiddleware(s.EditAddressHandler, utils.SellerRole))

	mux.HandleFunc("GET /seller/orders", middleware.AuthenticateUserMiddleware(s.GetOrdersHandler, utils.SellerRole))
	mux.HandleFunc("PUT /seller/orders/status", middleware.AuthenticateUserMiddleware(s.ChangeOrderStatusHandler, utils.SellerRole))
	mux.HandleFunc("GET /seller/sales/report", middleware.AuthenticateUserMiddleware(s.SalesReportHandler, utils.SellerRole))
}
