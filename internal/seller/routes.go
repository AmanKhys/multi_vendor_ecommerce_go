package seller

import (
	"net/http"
)

var s = Seller{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("post /seller/sign_up/", s.SignUpHandler)
	mux.HandleFunc("post /seller/sign_up_otp/", s.SignUpOTPHandler)
	mux.HandleFunc("post /seller/login/", s.LoginHandler)
	mux.HandleFunc("post /seller/sign_on_google/", s.SignUpOnGoogleHandler)

	mux.HandleFunc("get /seller/products/", s.ProductsHandler)
	mux.HandleFunc("get /seller/product/{productID}/details/", s.ProductDetailsHandler)
	mux.HandleFunc("post /seller/product/add", s.AddProductHandler)
	mux.HandleFunc("put /seller/product/{productID}/edit/", s.EditProductHandler)
	mux.HandleFunc("put /seller/product/{productID}/delete/", s.DeleteProductHandler)
}
