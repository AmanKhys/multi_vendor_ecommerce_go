package helpers

import (
	"net/http"

	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func GetUserHelper(w http.ResponseWriter, r *http.Request) *middleware.User {
	user, ok := r.Context().Value(utils.UserKey).(middleware.User)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return nil
	}
	return &user
}
