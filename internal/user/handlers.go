package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
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
	products, err := u.DB.GetAllProducts(context.TODO())
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
	ProductID := r.URL.Query().Get("id")
	id, err := uuid.Parse(ProductID)
	if err != nil {
		http.Error(w, "wrong productID format", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(context.TODO(), id)
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

func (u *User) CategoryHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	categoryName := queryParams.Get("category_name")

	products, err := u.DB.GetProductsByCategoryName(r.Context(), categoryName)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no available products by the cateogry:" + categoryName
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn("error fetching products by category name:", err.Error())
		http.Error(w, "internal error fetching products by category name", http.StatusBadRequest)
		return
	}

	var resp struct {
		Data    []db.GetProductsByCategoryNameRow `json:"data"`
		Message string                            `json:"message"`
	}
	resp.Data = products
	resp.Message = "successfully fetched products from category:" + categoryName
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) GetAddressesHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return
	}
	addresses, err := u.DB.GetAddressesByUserID(r.Context(), user.ID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no addresses added yet for the user."
		w.Write([]byte(message))
	} else if err != nil {
		http.Error(w, "internal server error fetching addresses by userID", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    []db.Address `json:"data"`
		Message string       `json:"message"`
	}
	resp.Data = addresses
	resp.Message = "successfully fetched all address of the user"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (u *User) AddAddressHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return
	} else if user.Role == AdminRole {
		http.Error(w, "admins can't add address", http.StatusUnauthorized)
		return
	}
	var arg db.AddAddressByUserIDParams
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, "invalid request data format", http.StatusBadRequest)
		return
	}
	arg.UserID = user.ID
	address, err := u.DB.AddAddressByUserID(context.TODO(), arg)
	if err != nil {
		log.Warn("error adding valid user address", err.Error())
		http.Error(w, "internal server error adding user address", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    db.Address `json:"address"`
		Message string     `json:"message"`
	}
	resp.Data = address
	resp.Message = "successfully added addres to the user with email:" + user.Email
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (u *User) EditAddressHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return
	}

	var req db.EditAddressByIDParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request format", http.StatusBadRequest)
		return
	} else if !(validators.ValidateAddress(req.BuildingName) &&
		validators.ValidateAddress(req.StreetName) &&
		validators.ValidateAddress(req.Town) &&
		validators.ValidateAddress(req.District) &&
		validators.ValidateAddress(req.State) &&
		validators.ValidatePincode(int(req.Pincode))) {
		http.Error(w, "invalid data values", http.StatusBadRequest)
		return
	}

	address, err := u.DB.GetAddressByID(context.TODO(), req.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid addressID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal error fetching AddressByID:", err.Error())
		http.Error(w, "internal error fetching address", http.StatusBadRequest)
		return
	} else if address.UserID != user.ID {
		http.Error(w, "not the current user's address to change; unauthorized", http.StatusUnauthorized)
		return
	}

	arg := req
	editedAddr, err := u.DB.EditAddressByID(context.TODO(), arg)
	if err != nil {
		log.Warn("internal error editing address for a valid address:", err.Error())
		http.Error(w, "internal error editing address for valid address", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    db.Address `json:"data"`
		Message string     `json:"message"`
	}
	resp.Data = editedAddr
	resp.Message = "successfully edited address"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) DeleteAddressHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("error fetching user from request context for user")
		http.Error(w, "internal server error marshalling user from request context.", http.StatusInternalServerError)
		return
	}

	var req struct {
		AddressID uuid.UUID `json:"address_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data format", http.StatusBadRequest)
		return
	}

	address, err := u.DB.GetAddressByID(context.TODO(), req.AddressID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid addressID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal server error fetching address by GetADdresByID:", err.Error())
		http.Error(w, "internal error fetching address from address_id", http.StatusInternalServerError)
		return
	} else if address.UserID != user.ID {
		http.Error(w, "not user's addres to delete. Unauthorized", http.StatusUnauthorized)
		return
	}
	err = u.DB.DeleteAddressByID(context.TODO(), req.AddressID)
	if err != nil {
		log.Warn("internal server error deleting valid address deleting request", err.Error())
		http.Error(w, "internal server error deleting valid address deleting request", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Message string `json:"message"`
	}
	resp.Message = "successfully deleted the address with id:" + address.ID.String()
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
