package seller

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var s = Seller{DB: DB}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /seller/products", s.ProductsHandler)
	mux.HandleFunc("GET /seller/product/{productID}/details", s.ProductDetailsHandler)
	mux.HandleFunc("POST /seller/product/add", s.AddProductHandler)
	mux.HandleFunc("PUT /seller/product/{productID}/edit", s.EditProductHandler)
	mux.HandleFunc("PUT /seller/product/{productID}/delete", s.DeleteProductHandler)
}
