package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Admin struct{ DB *db.Queries }

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

func (s *Admin) AuthenticateAdmin(w http.ResponseWriter, r *http.Request) (db.User, error) {
	sessionCookie, err := r.Cookie("SessionID")
	if err != nil {
		log.Warn("SessionID cookie not found")
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return db.User{}, errors.New("sessionID cookie missing")
	}
	if sessionCookie.Value == "" {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return db.User{}, errors.New("empty sessionID")
	}

	uid, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		log.Warn("Invalid sessionID format")
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return db.User{}, errors.New("invalid sessionID format")
	}

	user, err := s.DB.GetUserBySessionID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return db.User{}, errors.New("session not found")
		}
		log.Error("Database error fetching user by sessionID:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return db.User{}, errors.New("database error")
	}

	if user.Role != "admin" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return db.User{}, errors.New("user is not a customer")
	}

	return user, nil
}
