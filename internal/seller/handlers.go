package seller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Seller struct {
	DB *db.Queries
}

func (s *Seller) OwnProductsHandler(w http.ResponseWriter, r *http.Request) {
	var SessionID, err = r.Cookie("SessionID")
	if err != nil {
		log.Warn(err)
	}
	uid, err := uuid.Parse(SessionID.Value)
	if err != nil {
		log.Warn("error parsing sessionID", err)
		http.Error(w, "error parsing sessionID", http.StatusInternalServerError)
		return
	}
	user, err := s.DB.GetUserBySessionID(r.Context(), uid)
	if err != nil {
		log.Warn("error fetching Seller from sessionID:", SessionID)
		http.Error(w, "error fetching Seller from sessionID", http.StatusInternalServerError)
		return
	}
	products, err := s.DB.GetProductsBySellerID(r.Context(), user.ID)
	if err != nil {
		log.Warn("error fetching products for seller: ", user.ID, ":", err)
		http.Error(w, "unable to fetch seller products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	type Response struct {
		Data []db.Product `json:"data"`
	}
	resp := Response{Data: products}
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	var request req
	json.NewDecoder(r.Body).Decode(&request)
	product, err := s.DB.GetProductByID(r.Context(), request.ProductID)
	if err != nil {
		log.Warn("error fetching product from seller")
		http.Error(w, "error fetching product", http.StatusInternalServerError)
		return
	}
	type response struct {
		Data db.Product `json:"data"`
	}
	resp := response{Data: product}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Warn("error parsing response")
		http.Error(w, "error parsing response", http.StatusInternalServerError)
		return
	}
}

func (s *Seller) AddProductHandler(w http.ResponseWriter, r *http.Request) {
	var user, err = s.AuthenticateSeller(w, r)
	if err != nil {
		log.Warn(err)
		return
	}
	var arg db.AddProductParams
	json.NewDecoder(r.Body).Decode(&arg)
	arg.SellerID = user.ID
	product, err := s.DB.AddProduct(r.Context(), arg)
	if err != nil {
		log.Warnf("error adding product from sellerID: %s", user.ID)
		http.Error(w, "internal error while adding product", http.StatusInternalServerError)
		return
	}
	type resp struct {
		Data    db.Product `json:"data"`
		Message string     `json:"message"`
	}
	var response = resp{Data: product, Message: "product added successfully"}
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) EditProductHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.AuthenticateSeller(w, r)
	if err != nil {
		log.Warn(err)
	}
	var arg = db.EditProductByIDParams{}
	json.NewDecoder(r.Body).Decode(&arg)
	product, err := s.DB.EditProductByID(r.Context(), arg)
	if err == sql.ErrNoRows {
		http.Error(w, "no product with the specified id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn(err)
		http.Error(w, "error updating product details", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Data    db.Product `json:"data"`
		Message string     `json:"message"`
	}
	var response = resp{
		Data:    product,
		Message: "successfully updated product",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.AuthenticateSeller(w, r)
	if err != nil {
		log.Warn(err)
		return
	}
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	json.NewDecoder(r.Body).Decode(&req.ProductID)
	product, err := s.DB.DeleteProductByID(r.Context(), req.ProductID)
	if err != nil {
		log.Warn(err)
		http.Error(w, "error deleting product", http.StatusInternalServerError)
		return
	}
	type resp struct {
		Product db.Product `json:"product"`
		Message string     `json:"message"`
	}
	var response = resp{
		Product: product,
		Message: "successfully deleted product",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) AuthenticateSeller(w http.ResponseWriter, r *http.Request) (db.User, error) {
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

	if user.Role != "seller" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return db.User{}, errors.New("user is not a seller")
	}

	return user, nil
}
