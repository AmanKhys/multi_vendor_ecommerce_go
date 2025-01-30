package guest

import (
	"net/http"
)

var g = Guest{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", g.HomeHandler)
	mux.HandleFunc("get /login/", g.LoginPageHandler)
	mux.HandleFunc("post /login/", g.LoginHandler)
	mux.HandleFunc("get /signup/", g.SignUpPageHandler)
	mux.HandleFunc("post /signup/", g.SignUpHandler)
}
