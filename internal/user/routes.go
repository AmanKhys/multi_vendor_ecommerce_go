package user

import (
	"net/http"
)

var u = User{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("post /user/sign_up/", u.SignUpHandler)
	mux.HandleFunc("post /user/sign_up_otp", u.SignUpOTPHandler)
	mux.HandleFunc("post /user/login/", u.LoginHandler)
	mux.HandleFunc("post /user/sign_on_google/", u.SignOnGoogleHandler)

	mux.HandleFunc("get /user/products/", u.ProductsHandler)
	mux.HandleFunc("get /user/product/{productID}/", u.ProductHandler)
}
