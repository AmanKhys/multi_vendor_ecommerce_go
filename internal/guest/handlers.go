package guest

import (
	"context"
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
		log.Fatal("error decoding request json body: ", err)
		http.Error(w, "error decoding json request body", http.StatusBadRequest)
		return
	}

	var ctx context.Context
	u, err := g.DB.GetUserByEmail(ctx, user.Email)

	if err != nil {
		log.Fatal("error taking item from database: ", err)
		http.Error(w, "error fetching data from database", http.StatusInternalServerError)
		return
	}
	if u.Email == user.Email && u.Password == user.Password {
		w.Write([]byte("you've been logged in"))
		return
	} else if u.Email == user.Email {
		w.Write([]byte("wrong password"))
		return
	} else {
		w.Write([]byte("unable to login due to invalid fields"))
		return
	}

}

func (g *Guest) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello user you've singned up"))
}
