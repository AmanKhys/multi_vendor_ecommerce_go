package guest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/sessions"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	log "github.com/sirupsen/logrus"
)

var RoleSeller = "seller"
var RoleUser = "user"

type Guest struct {
	DB *db.Queries
}

func (g *Guest) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	message := "hello there"
	w.Write([]byte(message))
}

func (g *Guest) UserSignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req db.AddUserParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format:"+err.Error(), http.StatusBadRequest)
		return
	} else if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format:", http.StatusBadRequest)
		return
	} else if !validators.ValidateName(req.Name) {
		http.Error(w, "invalid Name format:", http.StatusBadRequest)
		return
	} else if !validators.ValidatePassword(req.Password) {
		http.Error(w, "invalid password format:", http.StatusBadRequest)
		return
	} else if req.Phone.Valid && !validators.ValidatePhone(strconv.Itoa(int(req.Phone.Int64))) {
		http.Error(w, "invalid phone format:", http.StatusBadRequest)
		return
	}
	user, _ := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if req.Email == user.Email && user.UserVerified {
		http.Error(w, "user already exists and verified", http.StatusBadRequest)
		return
	} else if req.Email == user.Email && !user.UserVerified {
		http.Error(w, "user already exists and not verified. visit /user_signup_otp and verify user", http.StatusBadRequest)
		return
	}

	log.Info("reached here mf")
	hashed, err := utils.HashPassword(req.Password)
	log.Info("reached here mf again")
	if err != nil {
		log.Warn("error hashing password")
		http.Error(w, "error hashing password"+err.Error(), http.StatusInternalServerError)
		return
	}
	var arg = db.AddUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
	}
	if req.Phone.Valid {
		arg.Phone = req.Phone
	}
	var respUser db.AddUserRow
	respUser, err = g.DB.AddUser(context.TODO(), arg)
	if err != nil {
		log.Warn("user not added")
		http.Error(w, "internal server error"+err.Error(), http.StatusInternalServerError)
		return
	}

	type resp struct {
		Data    db.AddUserRow `json:"data"`
		Message string        `json:"message"`
	}
	var response = resp{
		Data:    respUser,
		Message: "successfully added user. Now you need to verify it",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	signupOTP, err := g.DB.GetOTPByUserID(context.TODO(), respUser.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	}
	fmt.Println("testing otp: ", signupOTP)
}

func (g *Guest) SellerSignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req db.AddSellerParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format:"+err.Error(), http.StatusBadRequest)
		return
	} else if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format:", http.StatusBadRequest)
		return
	} else if !validators.ValidateName(req.Name) {
		http.Error(w, "invalid Name format:", http.StatusBadRequest)
		return
	} else if !validators.ValidatePassword(req.Password) {
		http.Error(w, "invalid password format:", http.StatusBadRequest)
		return
	} else if !validators.ValidatePhone(strconv.Itoa(int(req.Phone.Int64))) || !req.Phone.Valid {
		http.Error(w, "invalid phone format:", http.StatusBadRequest)
		return
	} else if !validators.ValidateGSTNo(req.GstNo.String) || !req.GstNo.Valid {
		http.Error(w, "invalid gst_no format:", http.StatusBadRequest)
		return
	} else if req.About.Valid && req.GstNo.String == "" {
		http.Error(w, "invalid about format: about empty", http.StatusBadRequest)
		return
	}
	user, _ := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if user.Role != "" && user.Role != RoleSeller {
		http.Error(w, "signing up on a already existing user. not allowed", http.StatusBadRequest)
		return
	}
	if req.Email == user.Email && user.EmailVerified {
		http.Error(w, "seller already exists and email verified", http.StatusBadRequest)
		return
	} else if req.Email == user.Email && !user.EmailVerified {
		http.Error(w, "seller already exists and email not verified. visit /user_signup_otp and verify user", http.StatusBadRequest)
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Warn("error hashing password")
		http.Error(w, "error hashing password"+err.Error(), http.StatusInternalServerError)
		return
	}
	var arg = db.AddSellerParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
		Phone:    req.Phone,
		GstNo:    req.GstNo,
		About:    req.About,
	}
	var respSeller db.AddSellerRow
	respSeller, err = g.DB.AddSeller(context.TODO(), arg)
	if err != nil {
		log.Warn("user not added")
		http.Error(w, "internal server error"+err.Error(), http.StatusInternalServerError)
		return
	}
	type resp struct {
		Data    db.AddSellerRow `json:"data"`
		Message string          `json:"message"`
	}
	var response = resp{
		Data:    respSeller,
		Message: "successfully added user. Now you need to verify it",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	signupOTP, err := g.DB.GetOTPByUserID(context.TODO(), respSeller.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	}
	fmt.Println("testing otp: ", signupOTP)
}

func (g *Guest) UserSignUpOTPHandler(w http.ResponseWriter, r *http.Request) {

}

func (g *Guest) SellerSignUpOTPHandler(w http.ResponseWriter, r *http.Request) {

}

func (g *Guest) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	user, err := g.DB.GetUserWithPasswordByEmail(context.TODO(), req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	err = utils.ComparePassword(req.Password, user.Password)
	log.Info(req.Password, user.Password)
	if err != nil {
		log.Warn(err)
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}

	var arg = db.AddSessionParams{
		UserID:    user.ID,
		IpAddress: utils.GetClientIPString(r),
		UserAgent: utils.GetUserAgent(r),
	}
	session, err := g.DB.AddSession(context.TODO(), arg)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	sessions.SetSessionCookie(w, session.ID.String())
	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("%s of id: %s has successfully logged in", user.Role, user.ID.String())
	w.Write([]byte(message))
	// http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (g *Guest) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := sessions.GetSessionCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessions.DeleteSessionCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
