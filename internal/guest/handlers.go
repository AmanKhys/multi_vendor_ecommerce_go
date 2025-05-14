package guest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	// check request validation
	var req db.AddUserParams
	var Err []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format:"+err.Error(), http.StatusBadRequest)
		return
	}
	if !validators.ValidateEmail(req.Email) {
		Err = append(Err, "invalid email format:")
	}
	if !validators.ValidateName(req.Name) {
		Err = append(Err, "invalid Name format:")
	}
	if !validators.ValidatePassword(req.Password) {
		Err = append(Err, "invalid password format:")
	}
	if req.Phone.Valid && !validators.ValidatePhone(strconv.Itoa(int(req.Phone.Int64))) {
		Err = append(Err, "invalid phone format:")
	}
	if len(Err) > 0 {
		http.Error(w, strings.Join(Err, "\n"), http.StatusBadRequest)
		return
	}

	// check if user already exists and is verified or not
	user, _ := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if req.Email == user.Email && user.UserVerified {
		http.Error(w, "user already exists and verified", http.StatusBadRequest)
		return
	} else if req.Email == user.Email && !user.UserVerified {
		http.Error(w, "user already exists and not verified. visit /user_signup_otp and verify user", http.StatusBadRequest)
		return
	}

	// hash password
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Warn("error hashing password")
		http.Error(w, "error hashing password"+err.Error(), http.StatusInternalServerError)
		return
	}

	// make addUserParams
	var arg = db.AddUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
	}
	if req.Phone.Valid {
		arg.Phone = req.Phone
	}

	// add user
	var addUser db.AddUserRow
	addUser, err = g.DB.AddUser(context.TODO(), arg)
	if err != nil {
		log.Warn("user not added")
		http.Error(w, "internal server error"+err.Error(), http.StatusInternalServerError)
		return
	}

	type respUser struct {
		ID    uuid.UUID     `json:"id"`
		Name  string        `json:"name"`
		Email string        `json:"email"`
		Phone sql.NullInt64 `json:"phone"`
		Role  string        `json:"role"`
	}
	var respUserData respUser = respUser{
		ID:    addUser.ID,
		Name:  addUser.Name,
		Email: addUser.Email,
		Phone: addUser.Phone,
		Role:  addUser.Role,
	}

	// give response
	var resp struct {
		Data    respUser `json:"data"`
		Message string   `json:"message"`
	}
	resp.Data = respUserData
	resp.Message = "Successfully added user. Now you need to verify it. Check email for otp."
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// make send otp to the newly added user to verify email
	signupOTP, err := g.DB.AddOTP(context.TODO(), addUser.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	}
	err = mail.SendOTPMail(int(signupOTP.Otp), signupOTP.ExpiresAt, addUser.Email)
	if err != nil {
		log.Warn("failed to send otp", err.Error())
		result, err := g.DB.DeleteOTPByEmail(context.TODO(), addUser.Email)
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
	// check req
	var req db.AddSellerParams
	// make Err slice for response strings
	var Err []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format:"+err.Error(), http.StatusBadRequest)
		return
	}
	if !validators.ValidateEmail(req.Email) {
		Err = append(Err, "invalid email format:")
	}
	if !validators.ValidateName(req.Name) {
		Err = append(Err, "invalid Name format:")
	}
	if !validators.ValidatePassword(req.Password) {
		Err = append(Err, "invalid password format:")
	}
	if !validators.ValidatePhone(strconv.Itoa(int(req.Phone.Int64))) || !req.Phone.Valid {
		Err = append(Err, "invalid phone format:")
	}
	if !validators.ValidateGSTNo(req.GstNo.String) || !req.GstNo.Valid {
		Err = append(Err, "invalid gst_no format:")
	}
	if req.About.Valid && req.About.String == "" {
		Err = append(Err, "invalid about format: about empty")
	}
	if len(Err) > 0 {
		http.Error(w, strings.Join(Err, "\n"), http.StatusBadRequest)
		return
	}

	// get seller and check if the seller exists, emailVerified, userVerified
	user, _ := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if user.Role != "" && user.Role != utils.SellerRole {
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

	// hash password for the newly creating seller
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Warn("error hashing password")
		http.Error(w, "error hashing password"+err.Error(), http.StatusInternalServerError)
		return
	}

	// craete arg for addSeller
	var arg = db.AddSellerParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
		Phone:    req.Phone,
		GstNo:    req.GstNo,
		About:    req.About,
	}
	var addSeller db.AddSellerRow
	// add seller
	addSeller, err = g.DB.AddSeller(context.TODO(), arg)
	if err != nil {
		log.Warn("user not added")
		http.Error(w, "internal server error"+err.Error(), http.StatusInternalServerError)
		return
	}

	type respSeller struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Phone int       `json:"phone"`
		Role  string    `json:"role"`
		GstNo string    `json:"gst_no"`
		About string    `json:"about"`
	}

	var respSellerData = respSeller{
		ID:    addSeller.ID,
		Name:  addSeller.Name,
		Email: addSeller.Email,
		Phone: int(addSeller.Phone.Int64),
		Role:  addSeller.Role,
		GstNo: addSeller.GstNo.String,
		About: addSeller.About.String,
	}

	// send response
	var resp struct {
		Data    respSeller `json:"data"`
		Message string     `json:"message"`
		Err     []string   `json:"errors"`
	}
	resp.Data = respSellerData
	resp.Message = "successfully added user. Now you need to verify it. Check email for otp."

	// send data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// make and send otp
	signupOTP, err := g.DB.AddOTP(context.TODO(), addSeller.ID)
	if err == sql.ErrNoRows {
		log.Warn("otp not generated")
		return
	} else if err != nil {
		log.Warn("error processing AddOTP query")
		return
	}

	err = mail.SendOTPMail(int(signupOTP.Otp), signupOTP.ExpiresAt, addSeller.Email)
	if err != nil {
		log.Warn("failed to send otp", err.Error())
		result, err := g.DB.DeleteOTPByEmail(context.TODO(), addSeller.Email)
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
	// make Err slice for response
	var validateErr []string
	json.NewDecoder(r.Body).Decode(&req)
	if !validators.ValidateEmail(req.Email) {
		validateErr = append(validateErr, "invalid email format:")
	}
	if !validators.ValidateOTP(req.Otp) {
		validateErr = append(validateErr, "invalid OTP format:")
	}
	if len(validateErr) > 0 {
		http.Error(w, strings.Join(validateErr, "\n"), http.StatusBadRequest)
		return
	}

	// check if the user exists
	user, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	} else if user.EmailVerified {
		http.Error(w, "user email already verified", http.StatusBadRequest)
		return
	}

	// chdck otp if the user exists and is not verified
	otp, err := g.DB.GetValidOTPByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		//make and  resend otp if there is no valid otp currently
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
			result, dbErr := g.DB.DeleteOTPByEmail(context.TODO(), req.Email)
			if dbErr != nil {
				log.Warn("internal error deleting otp by email after a failed attempt to send otp as email.")
			}
			k, err := result.RowsAffected()
			if err != nil {
				log.Warn("error fetching affected rows from db for deleted otp sql result")
			} else if k == 0 {
				log.Warn("no rows deleted after successful DeleteOTPByEmail query")
			}
			return
		}
		return

		// error handle the error to fetch the otp from the database the first time
	} else if err != nil {
		log.Warn("error fetching otp")
		http.Error(w, "internal server error fetching otp", http.StatusInternalServerError)
		return
	}

	// check if the otp from the reqest is correct
	if req.Otp != int(otp.Otp) {
		http.Error(w, "invalid otp", http.StatusBadRequest)
		return
	}

	// verify user
	verifiedUser, err := g.DB.VerifyUserByID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error verifying a valid user")
		http.Error(w, "internal server error verifying user", http.StatusInternalServerError)
		return
	}
	// errors and Messages [] slice for the response
	var Messages []string
	var Err []string
	// add a wallet for the user
	wallet, err := g.DB.AddWalletByUserID(context.TODO(), verifiedUser.ID)
	if err != nil {
		log.Warn("error adding wallet for newly created user:", err.Error())
		Err = append(Err, "unable to add wallet for the user after verifying. internal error")
	} else {
		Messages = append(Messages, "successfully added wallet for user:")
		Messages = append(Messages, "walletID:", wallet.ID.String(), fmt.Sprintf("savings: %v", wallet.Savings))

	}
	_, err = g.DB.DeleteOTPByEmail(context.TODO(), verifiedUser.Email)
	if err == sql.ErrNoRows {
		log.Warn("no otp deleted after executing query:", err.Error())
	} else if err != nil {
		log.Warn("error deleting otp after executing DeleteOTPByEmail query:", err)
	}

	type respUser struct {
		ID    uuid.UUID     `json:"id"`
		Name  string        `json:"name"`
		Email string        `json:"email"`
		Phone sql.NullInt64 `json:"phone"`
		Role  string        `json:"role"`
	}

	var respUserData = respUser{
		ID:    verifiedUser.ID,
		Name:  verifiedUser.Name,
		Email: verifiedUser.Email,
		Phone: verifiedUser.Phone,
		Role:  verifiedUser.Role,
	}
	// send response
	var resp struct {
		Data     respUser `json:"data"`
		Messages []string `json:"messages"`
		Err      []string `json:"errors"`
	}
	resp.Data = respUserData
	resp.Messages = append(Messages, "user verified successfully")
	resp.Err = Err

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (g *Guest) SellerSignUpOTPHandler(w http.ResponseWriter, r *http.Request) {
	// get req.Body and check if it's in correct format
	var req struct {
		Email string `json:"email"`
		Otp   int    `json:"otp"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json data request format", http.StatusBadRequest)
		return
	}

	// make Err slice for the response
	var validateErr []string
	if !validators.ValidateEmail(req.Email) {
		validateErr = append(validateErr, "invalid email format:")
	}
	if !validators.ValidateOTP(req.Otp) {
		validateErr = append(validateErr, "invalid OTP format:")
	}
	if len(validateErr) > 0 {
		http.Error(w, "invalid OTP format:", http.StatusBadRequest)
		return
	}

	// check if the user exists and is not verified
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
		// resend otp when there is no valid otp available
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
		return

		// error handle the first fetch for otp
	} else if err != nil {
		log.Warn("error fetching otp")
		http.Error(w, "internal server error fetching otp", http.StatusInternalServerError)
		return
	}

	// check if the otp is correct
	if req.Otp != int(otp.Otp) {
		http.Error(w, "invalid otp", http.StatusBadRequest)
		return
	}

	// verify user
	verifiedSeller, err := g.DB.VerifySellerEmailByID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error verifying a valid user")
		http.Error(w, "internal server error verifying user", http.StatusInternalServerError)
		return
	}
	type respSeller struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Phone int       `json:"phone"`
		Role  string    `json:"role"`
	}

	var respSellerData = respSeller{
		ID:    verifiedSeller.ID,
		Name:  verifiedSeller.Name,
		Email: verifiedSeller.Email,
		Phone: int(verifiedSeller.Phone.Int64),
		Role:  verifiedSeller.Role,
	}

	// send response
	var resp struct {
		Data    respSeller `json:"data"`
		Message string     `json:"message"`
	}
	resp.Data = respSellerData
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
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json data format in request", http.StatusBadRequest)
		return
	}
	// Err slice for error response for validation
	var validateErr []string
	if !validators.ValidateEmail(req.Email) {
		validateErr = append(validateErr, "invalid email format")
	}
	if !validators.ValidatePassword(req.Password) {
		validateErr = append(validateErr, "invalid password format")
	}
	if len(validateErr) > 0 {
		http.Error(w, strings.Join(validateErr, "\n"), http.StatusBadRequest)
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

	// make arg to create sessionID
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
	if err != nil {
		log.Warn("error deleting sessionByID")
		http.Error(w, "internal error terminating session", http.StatusInternalServerError)
		return
	}
}

func (g *Guest) DeleteSessionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// get session cookie
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

	// verify the sessionID from cookie
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

	// get userID from sessionID and delete all session instances
	userID := session.UserID
	_, err = g.DB.DeleteSessionsByuserID(context.TODO(), userID)
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

	// decode the userInfo.Body to our own Gauth jsonUserData var
	var jsonUserData Goauth
	if err = json.NewDecoder(UserInfo.Body).Decode(&jsonUserData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check if the user with the email already exists
	user, err := g.DB.GetUserByEmail(context.TODO(), jsonUserData.Email)
	if err == sql.ErrNoRows {
		// create a new user for the signed-in user and make a random hashed password for it then create the user
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

		// add user  arg
		var arg = db.AddAndVerifyUserParams{
			Name:     jsonUserData.Name,
			Email:    jsonUserData.Email,
			Password: hashedPassword,
		}
		// add user
		addUser, err := g.DB.AddAndVerifyUser(context.TODO(), arg)
		if err != nil {
			log.Warn("error adding user for google user:", err.Error())
			http.Error(w, "internal error adding user for google user.", http.StatusInternalServerError)
			return
		}

		// add session args
		var sessionArg = db.AddSessionParams{
			UserID:    addUser.ID,
			IpAddress: utils.GetClientIPString(r),
			UserAgent: utils.GetUserAgent(r),
		}
		// add session and set cookie
		session, err := g.DB.AddSession(context.TODO(), sessionArg)
		if err != nil {
			log.Warn("error adding session for google user:", err.Error())
			http.Error(w, "user added successfully. Unable to add session internal error", http.StatusOK)
			return
		}
		sessions.SetSessionCookie(w, session.ID.String())

		// make response
		var resp struct {
			Data    db.AddAndVerifyUserRow `json:"data"`
			Message string                 `json:"message"`
		}
		resp.Data = addUser
		resp.Message = "user has been successfully authenticated and logged in."
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		// handle failing to fetch and check if the user email already exists from the database
	} else if err != nil {
		log.Warn("internal error fetching user from GetUserByEmail and != sql.ErrNoRows and err != nil", err.Error())
		http.Error(w, "error fetching user data to check if the user already exists.", http.StatusInternalServerError)
		return
	}
	// check if the user is verified if the user was already registered. If not verify user and log
	if !user.EmailVerified {
		verifiedUser, err := g.DB.VerifyUserByID(context.TODO(), user.ID)
		if err != nil {
			log.Warn("internal server error verifying user on google authentication:", err.Error())
			http.Error(w, "internal server error  verifying user and setting sessionID cookie.", http.StatusInternalServerError)
			return
		}
		user.EmailVerified = verifiedUser.EmailVerified
		user.UserVerified = verifiedUser.UserVerified
	}

	// make add sessionArg
	var sessionArg = db.AddSessionParams{
		UserID:    user.ID,
		IpAddress: utils.GetClientIPString(r),
		UserAgent: utils.GetUserAgent(r),
	}
	// create session
	session, err := g.DB.AddSession(context.TODO(), sessionArg)
	if err != nil {
		log.Warn("unable to add goolge auth user session:", err.Error())
		http.Error(w, "unable to add google auth user session", http.StatusInternalServerError)
		return
	}
	sessions.SetSessionCookie(w, session.ID.String())

	type respUser struct {
		ID    uuid.UUID     `json:"id"`
		Name  string        `json:"name"`
		Email string        `json:"email"`
		Phone sql.NullInt64 `json:"phone"`
		Role  string        `json:"role"`
	}

	var respUserData = respUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  user.Role,
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	var resp struct {
		Data    respUser `json:"data"`
		Message string   `json:"message"`
	}
	resp.Data = respUserData
	resp.Message = "user logged in successfully and added session cookie."
	json.NewEncoder(w).Encode(resp)
}

func (g *Guest) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	// take req body and check if the fields are valid
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request body format", http.StatusBadRequest)
		return
	}
	if !validators.ValidateEmail(req.Email) {
		http.Error(w, "email in wrong format", http.StatusBadRequest)
		return
	}

	// get user and check if the user exists and is verified
	user, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid email; no user with this email", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Warn("error fetching user with email:", err.Error())
		http.Error(w, "internal error fetching user with email", http.StatusInternalServerError)
		return
	} else if !user.UserVerified {
		http.Error(w, "user not verified to change password", http.StatusBadRequest)
		return
	}

	// generate forgotOTP and send mail
	forgotOTP, err := g.DB.AddForgotOTPByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("internal error adding forgot_otp by userID:", err.Error())
		http.Error(w, "internal error adding forgot_otp by userID", http.StatusInternalServerError)
		return
	}
	err = mail.SendForgotOTPMail(int(forgotOTP.Otp), forgotOTP.ExpiresAt, user.Email)
	if err != nil {
		log.Warn("internal error sending email to valid user for forgotten password:", err.Error())
		http.Error(w, "internal error sending forgotten password otp mail", http.StatusInternalServerError)
		err := g.DB.DeleteForgotOTPByEmail(context.TODO(), user.Email)
		if err == sql.ErrNoRows {
			log.Warn("interna error no rows deleted in forgotOTP:", err.Error())
		} else if err != nil {
			log.Warn("internal error deleting forgotOTPByEmail:", err.Error())
		}
		return
	}

	// send response
	var resp struct {
		Message string `json:"message"`
	}
	resp.Message = fmt.Sprintf("successfully send otp mail to %s. Now send the email, otp and the new password in the url /forgot_otp", user.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (g *Guest) ForgotOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Otp      int    `json:"otp"`
		Password string `json:"password"`
	}

	// check if the req body and it's field are valid
	// make validateErr to give response if fields are not valid
	var validateErr []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body format", http.StatusBadRequest)
		return
	}
	if !validators.ValidateEmail(req.Email) {
		validateErr = append(validateErr, "invalid  email format")
	}
	if !validators.ValidateOTP(req.Otp) {
		validateErr = append(validateErr, "invalid otp format")
	}
	if !validators.ValidatePassword(req.Password) {
		validateErr = append(validateErr, "invalid password format")
	}
	if len(validateErr) > 0 {
		http.Error(w, strings.Join(validateErr, "\n"), http.StatusBadRequest)
		return
	}

	// get user and check
	user, err := g.DB.GetUserByEmail(context.TODO(), req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal error fetching user by email", http.StatusInternalServerError)
		return
	} else if !user.UserVerified {
		http.Error(w, "user not verified to change the password", http.StatusBadRequest)
		return
	}

	// take valid forgot otp from db
	forgotOtp, err := g.DB.GetValidForgotOTPByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		// make and send new valid otp if no valid forgot otp exists
		http.Error(w, "no valid forgot otps. generating a valid otp.", http.StatusBadRequest)
		addedOtp, err := g.DB.AddForgotOTPByUserID(context.TODO(), user.ID)
		if err != nil {
			log.Warn("error adding otp")
		}
		err = mail.SendForgotOTPMail(int(addedOtp.Otp), addedOtp.ExpiresAt, user.Email)
		if err != nil {
			log.Warn("error sending forgot otp mail:", err.Error())
			err := g.DB.DeleteForgotOTPByEmail(context.TODO(), user.Email)
			if err != nil {
				log.Warn("error deleting unsent forgot otp:", err.Error())
			}
		}

		// send response for the conveying to verify the newly created forgot otp
		var resp struct {
			Message string `json:"message"`
		}
		resp.Message = "successfully sent a forgot password otp to mail. try again with the new otp."
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return

		// deal with the fetching of valid otp
	} else if err != nil {
		log.Warn("internal error fetching valid forgotOtp from db", err.Error())
		http.Error(w, "internal server error fetching otp from db", http.StatusInternalServerError)
		return

		// if there is a valid otp check if the otp is the same as the request
	} else if int(forgotOtp.Otp) != req.Otp {
		http.Error(w, "incorrect otp", http.StatusBadRequest)
		return
	}

	// set password after confirming correct user and  valid otp
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Warn("intenral error hashing password while changing password")
		http.Error(w, "unable to hash and  change password. internal server error", http.StatusInternalServerError)
		return
	}

	// change password if the request otp is valid and correct
	var arg db.ChangePasswordByUserIDParams
	arg.ID = user.ID
	arg.Password = hash
	err = g.DB.ChangePasswordByUserID(context.TODO(), arg)
	if err != nil {
		log.Warn("internal error changing password in db")
		http.Error(w, "unable to save the hashed password in db. Changing password failed", http.StatusInternalServerError)
		return
	}

	// send response after successfully changing password
	var resp struct {
		Message string `json:"message"`
	}
	resp.Message = fmt.Sprintf("successfully changed the password of the user: %s with email: %s", user.Name, user.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
