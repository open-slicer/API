package db

import (
	"github.com/go-redis/redis"
	"slicerapi/internal/util"
)

// Redis is the Redis client.
var Redis *redis.Client

// Nil is Redis' Nil value. Used to avoid reimporting.
const Nil = redis.Nil

// Connect connects to the Redis server.
func Connect() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     util.Config.DB.Address,
		Password: util.Config.DB.Password,
		DB:       util.Config.DB.ID,
	})

	_, err := Redis.Ping().Result()
	util.Chk(err)
}
