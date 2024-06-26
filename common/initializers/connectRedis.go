package initializers

import (
	"context"
	"github.com/Arxtect/Einstein/config"
	"github.com/go-redis/redis/v8"
	"log"
)

var Rdb *redis.Client

func InitRedisClient(config *config.Config) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	_, err := Rdb.Ping(context.Background()).Result()

	if err != nil {
		log.Println("Redis connection failed", err)
		panic(err)
	}
}
