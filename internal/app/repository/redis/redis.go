package redis

import (
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/spf13/viper"
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

func Client() *redis.Client {
	return client
}
