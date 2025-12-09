package validators

import (
	"regexp"

	"github.com/google/uuid"
)

var (
	emailRegex          = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	gstNoRegex          = regexp.MustCompile(`^([0-3][0-9])([A-Z]{5}[0-9]{4}[A-Z])([1-9A-Z])Z([0-9A-Z])$`)
	nameRegex           = regexp.MustCompile(`^[a-zA-Z]{3,}[a-zA-Z ]*$`)
	hashedPasswordRegex = regexp.MustCompile(`^\$2[ayb]\$.{56}$`)
	passwordRegex       = regexp.MustCompile(`^[a-zA-Z0-9!@#$]{8,}$`)
	phoneRegex          = regexp.MustCompile(`^[1-9][0-9]{9}$`)
	roleRegex           = regexp.MustCompile(`^(user|seller|admin)$`)

	productNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,}[a-zA-Z0-9 ]*$`)

	addressRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,}[a-zA-Z0-9' ]*$`)

	couponNameRegex = regexp.MustCompile(`^[A-Z0-9]{3,}$`)

	reviewRatingRegex = regexp.MustCompile(`^[1-5]$`)
)

func ValidateUUIDStr(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidateGSTNo(gstNo string) bool {
	return gstNoRegex.MatchString(gstNo)
}

func ValidateName(name string) bool {
	return nameRegex.MatchString(name)
}

func ValidatePassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func ValidateHashedPassword(password string) bool {
	return hashedPasswordRegex.MatchString(password)
}

func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func ValidateRole(role string) bool {
	return roleRegex.MatchString(role)
}

func ValidateOTP(otp int) bool {
	flag := otp < 999999 && otp > 0
	return flag
}

// product validators
func ValidateProductName(name string) bool {
	return productNameRegex.MatchString(name)
}

func ValidateProductPrice(price float64) bool {
	return price > 0
}

func ValidateProductStock(stock int) bool {
	return stock >= 0
}

// address validators
func ValidateAddress(address string) bool {
	return addressRegex.MatchString(address)
}

func ValidatePincode(pincode int) bool {
	return (pincode >= 100000 && pincode <= 999999)
}

func ValidateCouponName(name string) bool {
	return couponNameRegex.MatchString(name)
}

func ValidateCouponPrice(price float64) bool {
	return price > 0
}

func ValidateReviewRating(rating string) bool {
	return reviewRatingRegex.MatchString(rating)
}
