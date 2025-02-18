package user

import (
	"net/http"

	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var u = User{DB: DB}

const UserRole = "user"
const SellerRole = "seller"
const AdminRole = "admin"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /user/products", u.ProductsHandler)
	mux.HandleFunc("GET /user/product", u.ProductHandler)
	mux.HandleFunc("GET /user/category", u.CategoryHandler)

	mux.HandleFunc("GET /user/address", middleware.AuthenticateUserMiddleware(u.GetAddressesHandler, UserRole))
	mux.HandleFunc("POST /user/address/add", middleware.AuthenticateUserMiddleware(u.AddAddressHandler, UserRole))
	mux.HandleFunc("PUT /user/address/edit", middleware.AuthenticateUserMiddleware(u.EditAddressHandler, UserRole))
	mux.HandleFunc("DELETE /user/address/delete", middleware.AuthenticateUserMiddleware(u.DeleteAddressHandler, UserRole))
}
