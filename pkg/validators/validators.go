package validators

import (
	"regexp"
)

var (
	emailRegex          = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	gstNoRegex          = regexp.MustCompile(`^([0-3][0-9])([A-Z]{5}[0-9]{4}[A-Z])([1-9A-Z])Z([0-9A-Z])$`)
	nameRegex           = regexp.MustCompile(`^[a-zA-Z]{3,}[a-zA-Z ]*$`)
	hashedPasswordRegex = regexp.MustCompile(`^\$2[ayb]\$.{56}$`)
	passwordRegex       = regexp.MustCompile(`^[a-zA-Z0-9!@#$]{8,}$`)
	phoneRegex          = regexp.MustCompile(`^[1-9][0-9]{9}$`)
	roleRegex           = regexp.MustCompile(`^(user|seller|admin)$`)
)

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
