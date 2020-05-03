package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/util"
)

const (
	BEARER_SCHEMA string = "Bearer "
)

// GetLoggedInUser extracts the auth token from req.
func GetLoggedInUser() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Grab the raw Authoirzation header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := util.ValidateToken(authHeader[len(BEARER_SCHEMA):])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			r.Header.Set("userID", claims.UserID)
			r.Header.Set("admin", strconv.FormatBool(claims.Admin))
			next.ServeHTTP(w, r)
		})
	}
}

func RequireUser() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("userID")
			if userID == "" {
				api.Respond(w, r, http.StatusUnauthorized, api.ErrUnauthorized)
				return
			}
			admin, _ := strconv.ParseBool(r.Header.Get("admin"))
			if admin == true {
				// Redirect to admin page if user is an admin.
				http.Redirect(w, r, "/admin", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			admin, err := strconv.ParseBool(r.Header.Get("admin"))
			if err != nil {
				api.Respond(w, r, http.StatusUnauthorized, api.ErrUnauthorized)
				return
			}
			if admin != true {
				api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
