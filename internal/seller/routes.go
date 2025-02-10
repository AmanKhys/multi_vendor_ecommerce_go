package seller

import (
	"net/http"

	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var s = Seller{DB: DB}

const SellerRole string = "seller"

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /seller/products", middleware.AuthenticateUserMiddleware(s.OwnProductsHandler, SellerRole))
	mux.HandleFunc("GET /seller/product", middleware.AuthenticateUserMiddleware(s.ProductDetailsHandler, SellerRole))
	mux.HandleFunc("POST /seller/product/add", middleware.AuthenticateUserMiddleware(s.AddProductHandler, SellerRole))
	mux.HandleFunc("PUT /seller/product/edit", middleware.AuthenticateUserMiddleware(s.EditProductHandler, SellerRole))
	mux.HandleFunc("DELETE /seller/product/delete", middleware.AuthenticateUserMiddleware(s.DeleteProductHandler, SellerRole))
}
