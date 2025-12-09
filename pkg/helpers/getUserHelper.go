package helpers

import (
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	log "github.com/sirupsen/logrus"
)

func (h *Helper) GetUserHelper(w http.ResponseWriter, r *http.Request) db.GetUserBySessionIDRow {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return db.GetUserBySessionIDRow{}
	}
	return user
}
