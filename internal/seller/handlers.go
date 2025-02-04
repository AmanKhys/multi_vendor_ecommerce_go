package seller

import (
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"net/http"
)

type Seller struct {
	DB *db.Queries
}

func (s *Seller) ProductsHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Seller) ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Seller) AddProductHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Seller) EditProductHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Seller) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {

}
