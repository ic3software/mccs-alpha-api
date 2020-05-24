package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/redis"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func RateLimiting() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "ratelimiting:" + util.IPAddress(r)

			val, err := redis.Client().Get(key).Result()
			if err != nil {
				err = redis.Client().Set(key, 1, viper.GetDuration("rate_limiting.duration")*time.Minute).Err()
				if err != nil {
					l.Logger.Error("[ERROR] redis Set key failed:", zap.Error(err))
				}
			}

			requestCount, _ := strconv.Atoi(val)
			if requestCount >= viper.GetInt("rate_limiting.limit") {
				api.Respond(w, r, http.StatusTooManyRequests)
				return
			}

			fmt.Println(requestCount)

			err = redis.Client().Incr(key).Err()
			if err != nil {
				l.Logger.Error("[ERROR] redis Incr key failed:", zap.Error(err))
			}

			next.ServeHTTP(w, r)
		})
	}
}
