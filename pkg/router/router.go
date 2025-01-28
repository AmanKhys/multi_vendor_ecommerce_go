package router

import (
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/admin"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/seller"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/user"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	seller.RegisterSellerRoutes(mux)
	user.RegisterUserRoutes(mux)
	admin.RegisterAdminRoutes(mux)

	return mux
}
