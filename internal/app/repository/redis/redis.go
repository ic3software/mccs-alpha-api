package redis

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/ic3network/mccs-alpha-api/util/l"
)

var (
	ctx                  = context.Background()
	client               *redis.Client
	rateLimitingDuration time.Duration
	loginAttemptsTimeout time.Duration
)

func init() {
	global.Init()

	// Get Redis configuration.
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	password := viper.GetString("redis.password")

	// Get other configuration values,
	rateLimitingDuration = viper.GetDuration(
		"rate_limiting.duration",
	) * time.Minute
	loginAttemptsTimeout = viper.GetDuration(
		"login_attempts.timeout",
	) * time.Second

	client = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0, // use default DB
	})

	waitForConnection()
}

func waitForConnection() {
	for {
		_, err := client.Ping(ctx).Result()
		if err != nil {
			log.Printf("Redis connection error: %+v \n", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
}

// GetRequestCount returns the request count for a given IP address.
func GetRequestCount(ip string) int {
	key := Ratelimiting + ":" + ip
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(val)
	return count
}

func IncRequestCount(ip string) error {
	key := Ratelimiting + ":" + ip
	_, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		// If the key does not exist, set it with a value of 1
		err = client.Set(
			ctx, key, 1, rateLimitingDuration,
		).Err()
	} else {
		err = client.Incr(ctx, key).Err()
	}

	if err != nil {
		l.Logger.Error("[ERROR] redis IncRequestCount failed:", zap.Error(err))
	}
	return err
}

// GetLoginAttempts returns the login attempts count for a given email address.
func GetLoginAttempts(email string) int {
	key := LoginAttempts + ":" + email
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(val)
	return count
}

func IncLoginAttempts(email string) error {
	key := LoginAttempts + ":" + email
	_, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		// If the key does not exist, set it with a value of 1.
		err = client.Set(
			ctx, key, 1, loginAttemptsTimeout,
		).Err()
	} else {
		err = client.Incr(ctx, key).Err()
	}

	if err != nil {
		l.Logger.Error("[ERROR] redis IncLoginAttempts failed:", zap.Error(err))
	}
	return err
}

// ResetLoginAttempts resets the login attempts count for a given email address.
func ResetLoginAttempts(email string) {
	key := LoginAttempts + ":" + email
	_, err := client.Del(ctx, key).Result()
	if err != nil {
		l.Logger.Error(
			"[ERROR] redis ResetLoginAttempts failed:",
			zap.Error(err),
		)
	}
}
