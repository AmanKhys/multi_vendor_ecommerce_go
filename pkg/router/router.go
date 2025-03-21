package router

import (
	"github.com/amankhys/multi_vendor_ecommerce_go/internal/admin"
	"github.com/amankhys/multi_vendor_ecommerce_go/internal/guest"
	"github.com/amankhys/multi_vendor_ecommerce_go/internal/seller"
	"github.com/amankhys/multi_vendor_ecommerce_go/internal/user"

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
