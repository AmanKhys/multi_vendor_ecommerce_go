package user

import (
	"net/http"
)

var u = User{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /user/sign_up/", u.SignUpHandler)
	mux.HandleFunc("POST /user/sign_up_otp", u.SignUpOTPHandler)
	mux.HandleFunc("POST /user/login/", u.LoginHandler)
	mux.HandleFunc("POST /user/sign_on_google/", u.SignOnGoogleHandler)

	mux.HandleFunc("GET /user/products/", u.ProductsHandler)
	mux.HandleFunc("GET /user/product/{productID}/", u.ProductHandler)
}
