package db

import (
	"github.com/go-redis/redis"
	"slicerapi/internal/util"
)

// Client is the Redis client.
var Client *redis.Client

// Nil is Redis' Nil value. Used to avoid reimporting.
const Nil = redis.Nil

// Connect connects to the Redis server.
func Connect() {
	Client = redis.NewClient(&redis.Options{
		Addr:     util.Config.DB.Address,
		Password: util.Config.DB.Password,
		DB:       util.Config.DB.ID,
	})

	_, err := Client.Ping().Result()
	util.Chk(err)
}
