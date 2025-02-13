package guest

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)
var g = Guest{DB: DB}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /home", g.HomeHandler)

	mux.HandleFunc("POST /login", g.LoginHandler)
	mux.HandleFunc("POST /user_signup", g.UserSignUpHandler)
	mux.HandleFunc("POST /seller_signup", g.SellerSignUpHandler)
	mux.HandleFunc("POST /user_signup_otp", g.UserSignUpOTPHandler)
	mux.HandleFunc("POST /seller_signup_otp", g.SellerSignUpOTPHandler)

	// mux.HandleFunc("GET /google_login", g.GoogleLoginPageHandler)
	// mux.HandleFunc("POST /google_login", g.GoogleLoginHandler)

	mux.HandleFunc("GET /logout", g.LogoutHandler)
	mux.HandleFunc("DELETE /delete_all_sessions", g.DeleteSessionHistoryHandler)
}
