package guest

import (
	"encoding/json"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/models/dto"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	log "github.com/sirupsen/logrus"
)

type Guest struct {
	DB *db.Queries
}

func (g *Guest) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there"))
}

func (g *Guest) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user dto.UserLoginParams
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Warn("error decoding request json body: ", err)
		http.Error(w, "error decoding json request body", http.StatusBadRequest)
		return
	}

	var ctx = r.Context()
	if user.Email == "" {
		http.Error(w, "email not specified in json data", http.StatusBadRequest)
		return
	}
	u, err := g.DB.GetUserByEmail(ctx, user.Email)
	if err != nil {
		log.Warn("error taking item from database: ", err)
		http.Error(w, "error fetching data from database", http.StatusInternalServerError)
		return
	}

	if u.Email == "" {
		http.Error(w, "no user found", http.StatusBadRequest)
		return
	} else if u.Email == user.Email && u.Password == user.Password {
		w.Write([]byte("you've been logged in"))
		return
	}
}

func (g *Guest) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello user you've singned up"))
}
