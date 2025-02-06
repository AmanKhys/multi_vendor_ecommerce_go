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
	// mux.HandleFunc("GET /login/", g.LoginPageHandler)
	mux.HandleFunc("POST /login", g.LoginHandler)
	// mux.HandleFunc("GET /signup/", g.SignUpPageHandler)
	mux.HandleFunc("POST /signup", g.SignUpHandler)

}
