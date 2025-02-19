package user

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
var u = User{DB: DB}
var helper = helpers.Helper{
	DB: DB,
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /user/products", u.ProductsHandler)
	mux.HandleFunc("GET /user/product", u.ProductHandler)
	mux.HandleFunc("GET /user/category", u.CategoryHandler)

	mux.HandleFunc("GET /user/address", middleware.AuthenticateUserMiddleware(u.GetAddressesHandler, utils.UserRole))
	mux.HandleFunc("POST /user/address/add", middleware.AuthenticateUserMiddleware(u.AddAddressHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/address/edit", middleware.AuthenticateUserMiddleware(u.EditAddressHandler, utils.UserRole))
	mux.HandleFunc("DELETE /user/address/delete", middleware.AuthenticateUserMiddleware(u.DeleteAddressHandler, utils.UserRole))

	mux.HandleFunc("GET /user/cart", middleware.AuthenticateUserMiddleware(u.GetCartHandler, utils.UserRole))
	mux.HandleFunc("POST /user/cart/add", middleware.AuthenticateUserMiddleware(u.AddCartHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/cart/edit", middleware.AuthenticateUserMiddleware(u.EditCartHandler, utils.UserRole))
	mux.HandleFunc("DELETE /user/cart/delete", middleware.AuthenticateUserMiddleware(u.DeleteCartHandler, utils.UserRole))

	mux.HandleFunc("GET /user/orders", middleware.AuthenticateUserMiddleware(u.GetOrdersHandler, utils.UserRole))
	mux.HandleFunc("POST /user/orders/addcart", middleware.AuthenticateUserMiddleware(u.AddCartToOrderHandler, utils.UserRole))
	mux.HandleFunc("POST /user/orders/add", middleware.AuthenticateUserMiddleware(u.AddProductToOrderHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/orders/cancel", middleware.AuthenticateUserMiddleware(u.CancelOrderHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/orders/return", middleware.AuthenticateUserMiddleware(u.ReturnOrderHandler, utils.UserRole))
}
