package seller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Seller struct {
	DB *db.Queries
}

func (s *Seller) OwnProductsHandler(w http.ResponseWriter, r *http.Request) {
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in reqeust context", http.StatusInternalServerError)
		return
	}
	products, err := s.DB.GetProductsBySellerID(context.TODO(), user.ID)
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
		ProductID uuid.UUID `json:"id"`
	}
	var request req
	json.NewDecoder(r.Body).Decode(&request)
	product, err := s.DB.GetProductByID(context.TODO(), request.ProductID)
	if err != nil {
		log.Warn("error fetching product from seller")
		http.Error(w, "error fetching product", http.StatusInternalServerError)
		return
	}
	type response struct {
		Data db.Product `json:"data"`
	}
	resp := response{Data: product}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Warn("error parsing response")
		http.Error(w, "error parsing response", http.StatusInternalServerError)
		return
	}
}

func (s *Seller) AddProductHandler(w http.ResponseWriter, r *http.Request) {
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
		return
	}
	var arg db.AddProductParams
	json.NewDecoder(r.Body).Decode(&arg)
	arg.SellerID = user.ID
	product, err := s.DB.AddProduct(context.TODO(), arg)
	if err != nil {
		log.Info("product", arg)
		log.Warnf("error adding product from sellerID: %s", user.ID)
		log.Warn(err)
		http.Error(w, "internal error while adding product", http.StatusInternalServerError)
		return
	}
	type resp struct {
		Data    db.Product `json:"data"`
		Message string     `json:"message"`
	}
	var response = resp{Data: product, Message: "product added successfully"}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) EditProductHandler(w http.ResponseWriter, r *http.Request) {
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
		return
	}
	var arg = db.EditProductByIDParams{}
	json.NewDecoder(r.Body).Decode(&arg)
	var productID = arg.ID
	seller, err := s.DB.GetSellerByProductID(context.TODO(), productID)

	if err == sql.ErrNoRows {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Warn("error fetching sellerID from database")
	}
	if seller.ID != user.ID {
		http.Error(w, "trying to edit products not owned by you", http.StatusBadRequest)
		return
	}

	// logic
	product, err := s.DB.EditProductByID(context.TODO(), arg)
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// take user from r context written by AuthenticateUserMiddleware
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
		return
	}

	// arguemnt struct to unmarshall from r.Body
	var req struct {
		ProductID uuid.UUID `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Warn(err)
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	// checking if the user.ID is the same as the product.SellerID
	var productID = req.ProductID
	seller, err := s.DB.GetSellerByProductID(context.TODO(), productID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Warn("error fetching sellerID from database")
	}
	if seller.ID != user.ID {
		http.Error(w, "trying to edit products not owned by you", http.StatusBadRequest)
		return
	}

	// business logic
	product, err := s.DB.DeleteProductByID(context.TODO(), req.ProductID)
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Seller) GetAllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := s.DB.GetAllCategories(context.TODO())
	if err != nil {
		log.Warn("error fetching all categories for seller:", err)
		http.Error(w, "internal error fetching cateogries", http.StatusInternalServerError)
		return
	}
	var resp struct {
		Data    []db.Category `json:"data"`
		Message string        `json:"message"`
	}
	resp.Data = categories
	resp.Message = "successfully fetched all categories"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) AddProductToCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("unable to fetch user from the request context after passing it from Auth Middleware")
		http.Error(w, "internal error fetching user by sessionID", http.StatusInternalServerError)
		return
	}
	var req struct {
		ProductID    uuid.UUID `json:"product_id"`
		CategoryName string    `json:"category_name"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request body format", http.StatusBadRequest)
		return
	}
	product, err := s.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching product from productID:", err.Error())
		http.Error(w, "internal server error fetching product from productID", http.StatusInternalServerError)
		return
	} else if product.SellerID != user.ID {
		http.Error(w, "user not authorized to add product to category, not user's product", http.StatusUnauthorized)
		return
	}
	category, err := s.DB.GetCategoryByName(context.TODO(), req.CategoryName)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching category by name query")
		http.Error(w, "internal error fetching category", http.StatusInternalServerError)
		return
	}

	var arg db.AddProductToCategoryByCategoryNameParams
	arg.CategoryName = req.CategoryName
	arg.ProductID = req.ProductID
	_, err = s.DB.AddProductToCategoryByCategoryName(context.TODO(), arg)
	if err != nil {
		log.Warn("error adding product to category items")
		http.Error(w, "internal error adding product to category items"+err.Error(), http.StatusInternalServerError)
		return
	}
	var resp struct {
		Product      db.Product `json:"product"`
		CategoryName string     `json:"category_name"`
		Message      string     `json:"message"`
	}
	resp.Product = product
	resp.CategoryName = category.Name
	resp.Message = "successfully added product to category items"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
