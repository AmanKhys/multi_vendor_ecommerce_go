package seller

import (
	"net/http"
)

func RegisterSellerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("post /seller/sign_up/", seller.SignUpHandler)
	mux.HandleFunc("post /seller/sign_up_otp/", seller.SignUPOTPHandler)
	mux.HandleFunc("post /seller/login/", seller.LoginHandler)
	mux.HandleFunc("post /seller/sign_on_google/", seller.SignUpOnGoogleHandler)

	mux.HandleFunc("get /seller/products/", seller.SellerProductsHandler)
	mux.HandleFunc("get /seller/product/{productID}/details/", seller.ProductDetailsHandler)
	mux.HandleFunc("get /seller/product/add", seller.AddProductHandler)
	mux.HandleFunc("update /seller/product/{productID}/edit/", seller.EditProductHandler)
	mux.HandleFunc("update /seller/product/{productID}/delete/", seller.DeleteProductHandler)
}
