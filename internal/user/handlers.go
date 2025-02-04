package user

import (
	// "encoding/json"
	// reqmod "github.com/amankhys/multi_vendor_ecommerce_go/models/request_model"
	// "github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"net/http"
)

type User struct {
	DB *db.Queries
}

func (u *User) ProductsHandler(w http.ResponseWriter, r *http.Request) {
}

func (u *User) ProductHandler(w http.ResponseWriter, r *http.Request) {

}
