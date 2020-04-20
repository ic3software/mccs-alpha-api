package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func CORS() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := cors.New(cors.Options{
				AllowCredentials: true,
			})
			c.HandlerFunc(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
