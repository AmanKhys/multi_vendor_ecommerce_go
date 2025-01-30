package admin

import (
	"net/http"
)

var a = Admin{}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("get /admin/users/", a.AdminUserHandler)
	mux.HandleFunc("put /admin/user/{userID}/block/", a.BlockUserHandler)
	mux.HandleFunc("put /admin/user/{userID}/unblock/", a.UnblockUserHandler)

	mux.HandleFunc("get /admin/sellers/", a.AdminSellerHandler)
	mux.HandleFunc("put /admin/seller/{sellerID}/block/", a.BlockSellerHandler)
	mux.HandleFunc("put /admin/seller/{sellerID}/unblock/", a.UnblockSellerHandler)

	mux.HandleFunc("get /admin/products/", a.AdminProductsHandler)
	mux.HandleFunc("delete /admin/product/{productID}/delete/", a.DeleteProductHandler)

	mux.HandleFunc("get /admin/categories/", a.AdminCategoriesHandler)
	mux.HandleFunc("post /admin/category/add/", a.AddCategoryHandler)
	mux.HandleFunc("put /admin/category/edit/", a.EditCategoryHandler)
	mux.HandleFunc("delete /admin/category/delete/", a.DeleteCategoryHandler)
}
