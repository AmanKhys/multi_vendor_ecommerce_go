package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Helper struct {
	DB *db.Queries
}

func (u *Helper) GetAddressesHelper(w http.ResponseWriter, r *http.Request, user db.GetUserBySessionIDRow) {
	addresses, err := u.DB.GetAddressesByUserID(r.Context(), user.ID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no addresses added yet for the user."
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn("internal server erro fetching address by userID:", err.Error())
		http.Error(w, "internal server error fetching addresses by userID", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    []db.GetAddressesByUserIDRow `json:"data"`
		Message string                       `json:"message"`
	}
	resp.Data = addresses
	resp.Message = "successfully fetched all address of the user"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (u *Helper) AddAddressHelper(w http.ResponseWriter, r *http.Request, user db.GetUserBySessionIDRow) {
	var req db.AddAddressByUserIDParams
	err := json.NewDecoder(r.Body).Decode(&req)
	var Err []string
	if err != nil {
		http.Error(w, "invalid request data format", http.StatusBadRequest)
		return
	}
	if !validators.ValidateAddress(req.BuildingName) {
		Err = append(Err, "invalid building name")
	}
	if !validators.ValidateAddress(req.StreetName) {
		Err = append(Err, "invalid street name")
	}
	if !validators.ValidateAddress(req.Town) {
		Err = append(Err, "invalid town name")
	}
	if !validators.ValidateAddress(req.District) {
		Err = append(Err, "invalid district name")
	}
	if !validators.ValidateAddress(req.State) {
		Err = append(Err, "invalid state name")
	}
	if !validators.ValidatePincode(int(req.Pincode)) {
		Err = append(Err, "invlaid pincode")
	}
	if len(Err) > 0 {
		http.Error(w, strings.Join(Err, "\n"), http.StatusBadRequest)
		return
	}
	req.UserID = user.ID
	req.Type = user.Role
	address, err := u.DB.AddAddressByUserID(context.TODO(), req)
	if err != nil {
		log.Warn("error adding valid user address", err.Error())
		http.Error(w, "internal server error adding user address", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    db.AddAddressByUserIDRow `json:"address"`
		Message string                   `json:"message"`
	}
	resp.Data = address
	resp.Message = "successfully added addres to the user with email:" + user.Email
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (u *Helper) EditAddressHelper(w http.ResponseWriter, r *http.Request, user db.GetUserBySessionIDRow) {
	var req db.EditAddressByIDParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request format", http.StatusBadRequest)
		return
	}
	// Error slice to store the return error values for response
	var Err []string
	if !validators.ValidateAddress(req.BuildingName) {
		Err = append(Err, "invalid building name")
	}
	if !validators.ValidateAddress(req.StreetName) {
		Err = append(Err, "invalid street name")
	}
	if !validators.ValidateAddress(req.Town) {
		Err = append(Err, "invalid town name")
	}
	if !validators.ValidateAddress(req.District) {
		Err = append(Err, "invalid district name")
	}
	if !validators.ValidateAddress(req.State) {
		Err = append(Err, "invalid state name")
	}
	if !validators.ValidatePincode(int(req.Pincode)) {
		Err = append(Err, "invlaid pincode")
	}
	if len(Err) > 0 {
		http.Error(w, strings.Join(Err, "\n"), http.StatusBadRequest)
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
		Data    db.EditAddressByIDRow `json:"data"`
		Message string                `json:"message"`
	}
	resp.Data = editedAddr
	resp.Message = "successfully edited address"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *Helper) DeleteAddressHelper(w http.ResponseWriter, r *http.Request, user db.GetUserBySessionIDRow) {
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
