package admin

import (
	"net/http"
)

type Admin struct{}

func (a *Admin) AdminUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) BlockUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) UnblockUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) AdminSellerHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) BlockSellerHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) UnblockSellerHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) AdminProductsHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) AddCategoryHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) EditCategoryHandler(w http.ResponseWriter, r *http.Request) {

}

func (a *Admin) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {

}
