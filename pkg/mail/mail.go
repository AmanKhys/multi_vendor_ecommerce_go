package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
)

func returnAuth() (smtp.Auth, error) {
	identity := os.Getenv(envname.SmtpIdentity)
	smtpHost := os.Getenv(envname.SmtpHost)
	smtpMail := os.Getenv(envname.SmtpEmail)
	smtpPassword := os.Getenv(envname.SmtpPassword)
	auth := smtp.PlainAuth(identity, smtpMail, smtpPassword, smtpHost)
	return auth, nil
}

func SendOTPMail(otp int, expires_at time.Time, recepientMail string) error {
	smtpServer := os.Getenv(envname.SmtpServer)
	smtpMail := os.Getenv(envname.SmtpEmail)

	auth, err := returnAuth()
	if err != nil {
		return err
	}
	var recepients []string
	message := fmt.Sprintf("From: %s\r\n", smtpMail) +
		"To: " + recepientMail + "\r\n" +
		"Subject: OTP Verification\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" + // Empty line separates headers from body
		fmt.Sprintf("Dear User,\n\n"+
			"Your One-Time Password (OTP) for verification is: %d\n\n"+
			"This OTP will expire on: %s\n"+
			"Time remaining: %s\n\n"+
			"Please do not share this OTP with anyone.\n\n"+
			"If you did not request this, please ignore this email.\n\n"+
			"Best regards,\nToy Stores Ecom",
			otp,
			expires_at.Format("Monday, 02 Jan 2006, 03:04 PM MST"), // More readable date format
			time.Until(expires_at).Round(time.Second).String())     // Time left rounded to seconds

	recepients = append(recepients, recepientMail)
	err = smtp.SendMail(smtpServer, auth, smtpMail, recepients, []byte(message))
	return err
}

func SendForgotOTPMail(otp int, expires_at time.Time, recepientMail string) error {
	smtpServer := os.Getenv(envname.SmtpServer)
	smtpMail := os.Getenv(envname.SmtpEmail)

	auth, err := returnAuth()
	if err != nil {
		return err
	}
	var recepients []string
	message := fmt.Sprintf("From: %s\r\n", smtpMail) +
		"To: " + recepientMail + "\r\n" +
		"Subject: Forgot Password OTP Verification\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" + // Empty line separates headers from body
		fmt.Sprintf("Dear User,\n\n"+
			"Your One-Time Password (OTP) to enter forget password is: %d\n\n"+
			"This OTP will expire on: %s\n"+
			"Time remaining: %s\n\n"+
			"Please do not share this OTP with anyone.\n\n"+
			"If you did not request this, please ignore this email.\n\n"+
			"Best regards,\nToy Stores Ecom",
			otp,
			expires_at.Format("Monday, 02 Jan 2006, 03:04 PM MST"), // More readable date format
			time.Until(expires_at).Round(time.Second).String())     // Time left rounded to seconds

	recepients = append(recepients, recepientMail)
	err = smtp.SendMail(smtpServer, auth, smtpMail, recepients, []byte(message))
	return err
}
