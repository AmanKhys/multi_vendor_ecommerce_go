package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	userpb "github.com/amankhys/multi_vendor_ecommerce_go/pkg/pb/user"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type User struct {
	ID    uuid.UUID
	Name  string
	Email string
	Role  string
	Phone string
}

func AuthenticateUserMiddleware(next http.HandlerFunc, role string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("SessionID")
		if err != nil {
			log.Warn("SessionID cookie not found")
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		if sessionCookie.Value == "" {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		uid, err := uuid.Parse(sessionCookie.Value)
		if err != nil {
			log.Warn("Invalid sessionID format")
			http.Error(w, "invalid session id format", http.StatusUnauthorized)
			return
		}

		// replace with grpc client
		userClient := userpb.NewUserServiceClient(&grpc.ClientConn{})
		user, err := userClient.GetUserBySessionID(context.TODO(), &userpb.GetUserBySessionIDRequest{SessionID: uid.String()})
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

		// set a gloabal user struct for the auth middleware
		contextUser := &User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
			Phone: user.Phone,
		}
		// Store user in context and call next handler
		ctx := context.WithValue(r.Context(), utils.UserKey, contextUser)
		next(w, r.WithContext(ctx))
	}
}
