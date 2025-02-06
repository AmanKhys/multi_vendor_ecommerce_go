package utils

import (
	"net/http"
)

// Write SessionID cookie function
func WriteCookie(w http.ResponseWriter, SessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "SessionID",
		Value:    SessionID,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
