package db

import (
	"slicerapi/internal/util"

	"github.com/go-redis/redis"
)

// Redis is the Redis client.
var Redis *redis.Client

// Nil is Redis' Nil value. Used to avoid reimporting.
const Nil = redis.Nil

// ConnectRedis connects to the Redis server.
func ConnectRedis() error {
	Redis = redis.NewClient(&redis.Options{
		Addr:     util.Config.DB.Redis.Address,
		Password: util.Config.DB.Redis.Password,
		DB:       util.Config.DB.Redis.DB,
	})

	_, err := Redis.Ping().Result()
	return err
}
