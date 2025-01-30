package router

import (
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/admin"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/guest"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/seller"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/internal/user"
	//	log "github.com/sirupsen/logrus"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	guest.RegisterRoutes(mux)
	user.RegisterRoutes(mux)
	seller.RegisterRoutes(mux)
	admin.RegisterRoutes(mux)

	return mux
}
