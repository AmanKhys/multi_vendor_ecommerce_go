package admin

import (
	"net/http"
)

func RegisterAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("get /admin/users/", admin.AdminUserHandler)
	mux.HandleFunc("update /admin/user/{userID}/block/", admin.BlockUserHandler)
	mux.HandleFunc("update /admin/user/{userID}/unblock/", admin.UnblockUserHandler)

	mux.HandleFunc("get /admin/sellers/", admin.AdminSellerHandler)
	mux.HandleFunc("update /admin/seller/{sellerID}/block/", admin.BlockSellerHandler)
	mux.HandleFunc("update /admin/seller/{sellerID}/unblock/", admin.UnblockSellerHandler)

	mux.HandleFunc("get /admin/products/", admin.AdminProductsHandler)
	mux.HandleFunc("update /admin/product/{productID}/delete/", admin.DeleteProductHandler)

	mux.HandleFunc("get /admin/categories/", admin.AdminCategoriesHandler)
	mux.HandleFunc("post /admin/category/add/", admin.AddCategoryHandler)
	mux.HandleFunc("update /admin/category/edit/", admin.EditCategoryHandler)
	mux.HandleFunc("update /admin/category/add/", admin.DeleteCategoryHandler)
}
