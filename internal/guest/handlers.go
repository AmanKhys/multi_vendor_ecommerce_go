package guest

import (
	"net/http"
)

type Guest struct {
}

func (g *Guest) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there mwonuse"))
}

func (g *Guest) LoginPageHandler(w http.ResponseWriter, r *http.Request) {

}

func (g *Guest) LoginHandler(w http.ResponseWriter, r *http.Request) {

}

func (g *Guest) SignUpPageHandler(w http.ResponseWriter, r *http.Request) {

}

func (g *Guest) SignUpHandler(w http.ResponseWriter, r *http.Request) {

}
