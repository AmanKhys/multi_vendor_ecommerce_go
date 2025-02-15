package guest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/mail"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/sessions"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var RoleSeller = "seller"
var RoleUser = "user"

type Guest struct {
	DB     *db.Queries
	config *oauth2.Config
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

	hashed, err := utils.HashPassword(req.Password)
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
	signupOTP, err := g.DB.AddOTP(context.TODO(), respUser.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	}
	err = mail.SendOTPMail(int(signupOTP.Otp), signupOTP.ExpiresAt, respUser.Email)
	if err != nil {
		log.Warn("failed to send otp", err.Error())
		result, err := g.DB.DeleteOTPByEmail(context.TODO(), respUser.Email)
		if err != nil {
			log.Warn("error deleting otp by email")
		}
		k, err := result.RowsAffected()
		if err != nil {
			log.Warn("error fetching the rows affected from DeleteOTPByEmail query resut")
		}
		if k == 0 {
			log.Warn("no rows affected while operating DeleteOTPByEmail query")
		}
	}
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
		Message: "successfully added user. Now you need to verify it. Check email for otp.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	signupOTP, err := g.DB.AddOTP(context.TODO(), respSeller.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	} else if err != nil {
		log.Warn("error processing AddOTP query")
		return
	}

	err = mail.SendOTPMail(int(signupOTP.Otp), signupOTP.ExpiresAt, respSeller.Email)
	if err != nil {
		log.Warn("failed to send otp", err.Error())
		result, err := g.DB.DeleteOTPByEmail(context.TODO(), respSeller.Email)
		if err != nil {
			log.Warn("error deleting otp by email")
		}
		k, err := result.RowsAffected()
		if err != nil {
			log.Warn("error fetching the rows affected from DeleteOTPByEmail query resut")
		}
		if k == 0 {
			log.Warn("no rows affected while operating DeleteOTPByEmail query")
		}
	}
}

func (g *Guest) UserSignUpOTPHandler(w http.ResponseWriter, r *http.Request) {
	// get req.Body and check if it's in correct format
	var req struct {
		Email string `json:"email"`
		Otp   int    `json:"otp"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format:", http.StatusBadRequest)
		return
	} else if !validators.ValidateOTP(req.Otp) {
		http.Error(w, "invalid OTP format:", http.StatusBadRequest)
		return
	}

	user, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	} else if user.EmailVerified {
		http.Error(w, "user email already verified", http.StatusBadRequest)
		return
	}

	otp, err := g.DB.GetValidOTPByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "no valid otp available. generating another otp", http.StatusBadRequest)
		otp, err := g.DB.AddOTP(context.TODO(), user.ID)
		if err == sql.ErrNoRows {
			log.Warn("no otp generated")
			return
		}
		err = mail.SendOTPMail(int(otp.Otp), otp.ExpiresAt, user.Email)
		if err != nil {
			log.Warn("error sending otp:", err.Error())
			http.Error(w, "error sending otp:", http.StatusInternalServerError)
			return
		}
		log.Info("testing otp generated: ", otp) // for testing
		return
	} else if err != nil {
		log.Warn("error fetching otp")
		http.Error(w, "internal server error fetching otp", http.StatusInternalServerError)
		return
	}
	if req.Otp != int(otp.Otp) {
		http.Error(w, "invalid otp", http.StatusBadRequest)
		return
	}

	// verify user
	respUser, err := g.DB.VerifyUserByID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error verifying a valid user")
		http.Error(w, "internal server error verifying user", http.StatusInternalServerError)
		return
	}
	_, err = g.DB.DeleteOTPByEmail(context.TODO(), respUser.Email)
	if err == sql.ErrNoRows {
		log.Warn("no otp deleted after executing query:", err.Error())
	} else if err != nil {
		log.Warn("error deleting otp after executing DeleteOTPByEmail query:", err)
	}

	// send response
	var resp struct {
		Data    db.VerifyUserByIDRow `json:"data"`
		Message string               `json:"message"`
	}
	resp.Data = respUser
	resp.Message = "user verified successfully"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (g *Guest) SellerSignUpOTPHandler(w http.ResponseWriter, r *http.Request) {
	// get req.Body and check if it's in correct format
	var req struct {
		Email string `json:"email"`
		Otp   int    `json:"otp"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format:", http.StatusBadRequest)
		return
	} else if !validators.ValidateOTP(req.Otp) {
		http.Error(w, "invalid OTP format:", http.StatusBadRequest)
		return
	}

	user, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	} else if user.EmailVerified {
		http.Error(w, "user email already verified", http.StatusBadRequest)
		return
	}

	// validate otp
	otp, err := g.DB.GetValidOTPByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "no valid otp available. generating another otp", http.StatusBadRequest)
		otp, err := g.DB.AddOTP(context.TODO(), user.ID)
		if err == sql.ErrNoRows {
			log.Warn("no otp generated")
			return
		}
		err = mail.SendOTPMail(int(otp.Otp), otp.ExpiresAt, user.Email)
		if err != nil {
			result, err := g.DB.DeleteOTPByEmail(context.TODO(), user.Email)
			if err != nil {
				log.Warn("error deleting otp:", err)
			}
			k, err := result.RowsAffected()
			if err != nil {
				log.Warn("error fetching rows affected from sql.Result:", err)
			} else if k == 0 {
				log.Warn("no otp deleted after successful query execution")
			}
			log.Warn("error sending otp email to seller", err.Error())
			http.Error(w, "error sending otp email", http.StatusInternalServerError)
			return
		}
		log.Info("testing otp generated: ", otp) // for testing
		return
	} else if err != nil {
		log.Warn("error fetching otp")
		http.Error(w, "internal server error fetching otp", http.StatusInternalServerError)
		return
	}
	if req.Otp != int(otp.Otp) {
		http.Error(w, "invalid otp", http.StatusBadRequest)
		return
	}

	// verify user
	respSeller, err := g.DB.VerifySellerEmailByID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error verifying a valid user")
		http.Error(w, "internal server error verifying user", http.StatusInternalServerError)
		return
	}

	// send response
	var resp struct {
		Data    db.VerifySellerEmailByIDRow `json:"data"`
		Message string                      `json:"message"`
	}
	resp.Data = respSeller
	resp.Message = "seller email verified successfully. Wait for admin to verify the seller."

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (g *Guest) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// take request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	} else if !validators.ValidatePassword(req.Password) {
		http.Error(w, "invalid password format", http.StatusBadRequest)
		return
	}

	// compare email and password
	user, err := g.DB.GetUserWithPasswordByEmail(context.TODO(), req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusUnauthorized)
		return
	}

	err = utils.ComparePassword(req.Password, user.Password)
	if err != nil {
		log.Warn(err)
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	} else if !user.UserVerified {
		message := fmt.Sprintf("%s not verified", user.Role)
		http.Error(w, message, http.StatusUnauthorized)
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
	message := fmt.Sprintf("%s of id: %s has successfully logged in\n", user.Role, user.ID.String())
	w.Write([]byte(message))
}

func (g *Guest) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := sessions.GetSessionCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessions.DeleteSessionCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	id, err := uuid.Parse(cookie.Name)
	if err != nil {
		http.Error(w, "sessionID not in a valid format.", http.StatusBadRequest)
		return
	}
	// didn't check the result of query since there is only one matching id
	// no need to check if it is affected for more than one rows
	_, err = g.DB.DeleteSessionByID(context.TODO(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid sessionID in session cookie. unable to terminate a non-existing session", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error deleting sessionByID")
		http.Error(w, "internal error terminating session", http.StatusInternalServerError)
		return
	}
}

func (g *Guest) DeleteSessionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := sessions.GetSessionCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(cookie.Name)
	if err != nil {
		http.Error(w, "sessionID not in a valid format.", http.StatusBadRequest)
		return
	}
	session, err := g.DB.GetSessionDetailsByID(context.TODO(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid sessionID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal error deleting session row by id")
		http.Error(w, "internal error termainating session", http.StatusInternalServerError)
		return
	}
	if time.Now().Compare(session.ExpiresAt) == -1 {
		http.Error(w, "session already expired. Only authorized for a valid session.", http.StatusUnauthorized)
		return
	}
	userID := session.UserID
	_, err = g.DB.DeleteSessionsByuserID(context.TODO(), userID)
	// no need to check errNoRows. since I would atleast have one session on
	// if I am to get a session and the userID from it; theoretically impossible
	if err != nil {
		log.Warn("intenral error deleting sessions by userID in DeleteSessionHistoryHandler")
		http.Error(w, "internal server error deleting all session history", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	message := "successfully deleted all sessions history for the user"
	w.Write([]byte(message))
}

func (g *Guest) OauthHandler(w http.ResponseWriter, r *http.Request) {
	url := g.config.AuthCodeURL("helloo world", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (g *Guest) OauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// Exchanging the code for an access token
	t, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		log.Warn("internal error fetching response from the googleapi")
		http.Error(w, "internal error fetching response from the googleapi", http.StatusInternalServerError)
		return
	}

	// Creating an HTTP client to make authenticated request using the access key.
	// This client method also regenerate the access key using the refresh key.
	client := g.config.Client(context.Background(), t)

	// Getting the user public details from google API endpoint
	UserInfo, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Closing the request body when this function returns.
	// This is a good practice to avoid memory leak
	defer UserInfo.Body.Close()

	type Goauth struct {
		Id            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locate        string `json:"locale"`
	}

	var jsonUserData Goauth
	if err = json.NewDecoder(UserInfo.Body).Decode(&jsonUserData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := g.DB.GetUserByEmail(context.TODO(), jsonUserData.Email)
	if err == sql.ErrNoRows {
		password, err := utils.GenerateRandomString(16)
		if err != nil {
			log.Warn("error producing random string for google password:", err.Error())
			http.Error(w, "intenral error producing password for google user.", http.StatusInternalServerError)
			return
		}
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			log.Warn("intenral error hashing password for google user")
			http.Error(w, "internal error producing password for google user.", http.StatusInternalServerError)
			return
		}
		var arg = db.AddAndVerifyUserParams{
			Name:     jsonUserData.Name,
			Email:    jsonUserData.Email,
			Password: hashedPassword,
		}
		sessionUser, err := g.DB.AddAndVerifyUser(context.TODO(), arg)
		if err != nil {
			log.Warn("error adding user for google user:", err.Error())
			http.Error(w, "internal error adding user for google user.", http.StatusInternalServerError)
			return
		}

		var sessionArg = db.AddSessionParams{
			UserID:    sessionUser.ID,
			IpAddress: utils.GetClientIPString(r),
			UserAgent: utils.GetUserAgent(r),
		}
		session, err := g.DB.AddSession(context.TODO(), sessionArg)
		if err != nil {
			log.Warn("error adding session for google user:", err.Error())
			http.Error(w, "user added successfully. Unable to add session internal error", http.StatusOK)
			return
		}
		sessions.SetSessionCookie(w, session.ID.String())

		var resp struct {
			Data    db.AddAndVerifyUserRow `json:"data"`
			Message string                 `json:"message"`
		}
		resp.Data = sessionUser
		resp.Message = "user has been successfully authenticated and logged in."
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else if err != nil {
		log.Warn("internal error fetching user from GetUserByEmail and != sql.ErrNoRows and err != nil", err.Error())
		http.Error(w, "error fetching user data to check if the user already exists.", http.StatusInternalServerError)
		return
	}
	if !user.EmailVerified {
		verifiedUser, err := g.DB.VerifyUserByID(context.TODO(), user.ID)
		if err != nil {
			log.Warn("internal server error verifying user on google authentication:", err.Error())
			http.Error(w, "internal server error  verifying user and setting sessionID cookie.", http.StatusInternalServerError)
			return
		}
		user.EmailVerified = verifiedUser.EmailVerified
		user.UserVerified = verifiedUser.UserVerified
		user.UpdatedAt = verifiedUser.UpdatedAt
	}

	var sessionArg = db.AddSessionParams{
		UserID:    user.ID,
		IpAddress: utils.GetClientIPString(r),
		UserAgent: utils.GetUserAgent(r),
	}
	session, err := g.DB.AddSession(context.TODO(), sessionArg)
	if err != nil {
		log.Warn("unable to add goolge auth user session:", err.Error())
		http.Error(w, "unable to add google auth user session", http.StatusInternalServerError)
		return
	}
	sessions.SetSessionCookie(w, session.ID.String())
	w.Header().Set("Content-Type", "application/json")
	var resp struct {
		Data    db.GetUserByEmailRow `json:"data"`
		Message string               `json:"message"`
	}
	resp.Data = user
	resp.Message = "user logged in successfully and added session cookie."
	json.NewEncoder(w).Encode(resp)
}
