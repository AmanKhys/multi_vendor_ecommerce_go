package user

import (
	"net/http"
)

func RegisterUserHandlers(mux *http.ServeMux) {
	mux.HandleFunc("post /user/sign_up/", user.SignUpHandler)
	mux.HandleFunc("post /user/sign_up_otp", user.SignUpOTPHandler)
	mux.HandleFunc("post /user/login/", user.LoginHandler)
	mux.HandleFunc("post /user/sign_on_google/", user.SignOnGoogleHandler)

	mux.HandleFunc("get /user/products/", user.ProductsHandler)
	mux.HandleFunc("get /user/product/{productID}/", user.ProductHandler)
	mux.HandleFunc("get /user/product/{productID}/reviews/", user.ProductReviewHandler)
}
