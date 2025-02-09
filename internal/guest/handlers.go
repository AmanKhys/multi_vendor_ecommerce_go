package guest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/sessions"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	log "github.com/sirupsen/logrus"
)

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
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	// check if the user already exists and is verified
	u, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if u.Email != "" && u.EmailVerified == false {
		http.Redirect(w, r, "/user_signup_otp", http.StatusSeeOther)
		return
	} else if u.Email != "" && u.UserVerified == true {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}
	user, err := g.DB.AddUser(context.TODO(), req)
	if err != nil {
		log.Warn("error adding user.")
		http.Error(w, "internal error to add user", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Data    db.AddUserRow `json:"data"`
		Message string        `json:"message"`
	}
	var response = resp{
		Data:    user,
		Message: "successfully added user. Now you need to verify it",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (g *Guest) SellerSignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req db.AddSellerParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	// check if the seller already exists and is verified
	u, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if u.Email != "" && u.EmailVerified == false {
		http.Redirect(w, r, "/seller_signup_otp", http.StatusSeeOther)
		return
	} else if u.Email != "" && u.EmailVerified == true && u.UserVerified == false {
		w.Header().Set("Content-Type", "text/plain")
		message := "user email verified. Waiting for admin to admit seller as a verified seller."
		w.Write([]byte(message))
		return
	} else if u.Email != "" && u.UserVerified == true {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	user, err := g.DB.AddSeller(context.TODO(), req)
	if err != nil {
		log.Warn("error adding user.")
		http.Error(w, "internal error to add seller", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Data    db.AddSellerRow `json:"data"`
		Message string          `json:"message"`
	}
	var response = resp{
		Data:    user,
		Message: "successfully added user. Now you need to verify it",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
