package seller

import (
	"net/http"
)

var s = Seller{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /seller/sign_up/", s.SignUpHandler)
	mux.HandleFunc("POST /seller/sign_up_otp/", s.SignUpOTPHandler)
	mux.HandleFunc("POST /seller/login/", s.LoginHandler)
	mux.HandleFunc("POST /seller/sign_on_google/", s.SignUpOnGoogleHandler)

	mux.HandleFunc("GET /seller/products/", s.ProductsHandler)
	mux.HandleFunc("GET /seller/product/{productID}/details/", s.ProductDetailsHandler)
	mux.HandleFunc("POST /seller/product/add", s.AddProductHandler)
	mux.HandleFunc("PUT /seller/product/{productID}/edit/", s.EditProductHandler)
	mux.HandleFunc("PUT /seller/product/{productID}/delete/", s.DeleteProductHandler)
}
