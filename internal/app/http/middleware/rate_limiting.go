package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/redis"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
)

func RateLimiting() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := util.IPAddress(r)

			requestCount := redis.GetRequestCount(ip)

			if requestCount >= viper.GetInt("rate_limiting.limit") {
				api.Respond(w, r, http.StatusTooManyRequests)
				return
			}

			redis.IncRequestCount(ip)

			next.ServeHTTP(w, r)
		})
	}
}
