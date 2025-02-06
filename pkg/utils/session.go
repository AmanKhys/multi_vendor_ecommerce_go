package utils

import (
	"net/http"

	log "github.com/sirupsen/logrus"
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

// Reead SessionID cookie function
func ReadCookie(r *http.Request, name string) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)

	if cookie == nil {
		return nil, http.ErrNoCookie
	}
	if err != nil {
		log.Warnf("error fetching cookie %s", name)
		http.Error(w, "error fetching cookie", http.StatusInternalServerError)
		return
	}

	return cookie, nil
}
