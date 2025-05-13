package guest

import (
	"net/http"
	"os"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// get db connection and get the *db.Queries()
var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)

// load the env for clientID and clientSecret for googelAuthConfig
var clientId = os.Getenv(envname.GoogleClientID)
var clientSecret = os.Getenv(envname.GoogleSecretKey)

// set the config for google auth and give it the guest struct object
var conf = &oauth2.Config{
	ClientID:     clientId,
	ClientSecret: clientSecret,
	RedirectURL:  "http://localhost:7777/auth/callback",
	Scopes:       []string{"email", "profile"},
	Endpoint:     google.Endpoint,
}

// set the guest object from the guest Struct
var g = Guest{
	DB:     DB,
	config: conf,
}

// register the routes for the guest
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
