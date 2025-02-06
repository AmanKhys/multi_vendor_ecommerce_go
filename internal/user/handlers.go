package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type User struct {
	DB *db.Queries
}

func (u *User) ProductsHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Data []db.Product `json:"data"`
	}
	products, err := u.DB.GetAllProducts(r.Context())
	if err != nil {
		log.Warn("couldn't fetch from products table: ", err)
		http.Error(w, "internal server error: couldn't fetch products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Data: products}
	json.NewEncoder(w).Encode(resp)
}

func (u *User) ProductHandler(w http.ResponseWriter, r *http.Request) {
	ProductID := r.PathValue("productID")
	id, err := uuid.Parse(ProductID)
	if err != nil {
		http.Error(w, "wrong productID format", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(r.Context(), id)
	if err != nil {
		http.Error(w, "no such product exists", http.StatusNotFound)
		return
	}
	type Response struct {
		Data db.Product `json:"data"`
	}
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Data: product}
	json.NewEncoder(w).Encode(resp)
}

func (s *User) AuthenticateUser(w http.ResponseWriter, r *http.Request) (db.User, error) {
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

	if user.Role != "user" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return db.User{}, errors.New("user is not a customer")
	}

	return user, nil
}
