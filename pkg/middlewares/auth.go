package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)

func AuthenticateUserMiddleware(next http.HandlerFunc, role string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("SessionID")
		if err != nil {
			log.Warn("SessionID cookie not found")
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		log.Info("Received SessionID:", sessionCookie.Value)

		if sessionCookie.Value == "" {
			log.Info("Received SessionID : ", sessionCookie.Value)
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		uid, err := uuid.Parse(sessionCookie.Value)
		if err != nil {
			log.Warn("Invalid sessionID format")
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		user, err := DB.GetUserBySessionID(r.Context(), uid)
		log.Info(user)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}
			log.Error("Database error fetching user by sessionID:", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Check role if needed
		if user.Role != role {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		// Store user in context and call next handler
		ctx := context.WithValue(r.Context(), utils.UserKey, user)
		next(w, r.WithContext(ctx))
	}
}
