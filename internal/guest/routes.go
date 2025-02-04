package guest

import (
	"net/http"
)

var g = Guest{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", g.HomeHandler)
	mux.HandleFunc("GET /login/", g.LoginPageHandler)
	mux.HandleFunc("POST /login/", g.LoginHandler)
	mux.HandleFunc("GET /signup/", g.SignUpPageHandler)
	mux.HandleFunc("POST /signup/", g.SignUpHandler)
}
