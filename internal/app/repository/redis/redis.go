package redis

import (
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var client *redis.Client

func init() {
	global.Init()

	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	password := viper.GetString("redis.password")
	client = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0, // use default DB
	})

	for {
		_, err := client.Ping().Result()
		if err != nil {
			log.Printf("Redis connection error: %+v \n", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
}

func GetRequestCount(ip string) int {
	key := Ratelimiting + ":" + ip
	val, err := client.Get(key).Result()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(val)
	return count
}

func IncRequestCount(ip string) error {
	key := Ratelimiting + ":" + ip

	_, err := client.Get(key).Result()
	if err == redis.Nil {
		err = client.Set(key, 1, viper.GetDuration("rate_limiting.duration")*time.Minute).Err()
		if err != nil {
			l.Logger.Error("[ERROR] redis IncRequestCount failed:", zap.Error(err))
			return err
		}
	} else {
		err = client.Incr(key).Err()
		if err != nil {
			l.Logger.Error("[ERROR] redis IncRequestCount failed:", zap.Error(err))
			return err
		}
	}

	return nil
}

func GetLoginAttempts(email string) int {
	key := LoginAttempts + ":" + email
	val, err := client.Get(key).Result()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(val)
	return count
}

func IncLoginAttempts(email string) error {
	key := LoginAttempts + ":" + email

	_, err := client.Get(key).Result()
	if err == redis.Nil {
		err = client.Set(key, 1, viper.GetDuration("login_attempts.timeout")*time.Second).Err()
		if err != nil {
			l.Logger.Error("[ERROR] redis IncLoginAttempts failed:", zap.Error(err))
			return err
		}
	} else {
		err = client.Incr(key).Err()
		if err != nil {
			l.Logger.Error("[ERROR] redis IncLoginAttempts failed:", zap.Error(err))
			return err
		}
	}

	return nil
}

func ResetLoginAttempts(email string) {
	key := LoginAttempts + ":" + email
	_, err := client.Del(key).Result()
	if err != nil {
		l.Logger.Error("[ERROR] redis ResetLoginAttempts failed:", zap.Error(err))
	}
}
