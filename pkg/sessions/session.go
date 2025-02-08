package sessions

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

const sessionCookieName string = "session_id"

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	var cookie = &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Expires:  time.Now().Add(7 * time.Hour * 24),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

func DeleteSessionCookie(w http.ResponseWriter, cookie *http.Cookie) {
	cookie.Expires = time.Unix(0, 0)
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(sessionCookieName)
}

func GenerateSessionID() uuid.UUID {
	return uuid.New()
}
