package guest

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	env "github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)

var envM, _ = env.Read(".env")
var clientId = envM[envname.GoogleClientID]
var clientSecret = envM[envname.GoogleSecretKey]

var conf = &oauth2.Config{
	ClientID:     clientId,
	ClientSecret: clientSecret,
	RedirectURL:  "http://localhost:7777/auth/callback",
	Scopes:       []string{"email", "profile"},
	Endpoint:     google.Endpoint,
}

var g = Guest{
	DB:     DB,
	config: conf,
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /home", g.HomeHandler)

	mux.HandleFunc("POST /login", g.LoginHandler)
	mux.HandleFunc("POST /user_signup", g.UserSignUpHandler)
	mux.HandleFunc("POST /seller_signup", g.SellerSignUpHandler)
	mux.HandleFunc("POST /user_signup_otp", g.UserSignUpOTPHandler)
	mux.HandleFunc("POST /seller_signup_otp", g.SellerSignUpOTPHandler)

	mux.HandleFunc("POST /forgot_password", g.ForgotPasswordHandler)
	mux.HandleFunc("POST /forgot_otp", g.ForgotOTPHandler)

	mux.HandleFunc("GET /auth/login", g.OauthHandler)
	mux.HandleFunc("GET /auth/callback", g.OauthCallbackHandler)

	mux.HandleFunc("GET /logout", g.LogoutHandler)
	mux.HandleFunc("DELETE /delete_all_sessions", g.DeleteSessionHistoryHandler)
}
