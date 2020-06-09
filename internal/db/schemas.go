package db

import (
	"time"
)

type Channel struct {
	ID      string          `bson:"_id"`
	Name    string          `bson:"name"`
	Date    time.Time       `bson:"date"`
	Pending map[string]bool `bson:"pending"`
	Users   map[string]bool `bson:"users"`
	Parent  string          `bson:"parent"`
}

type User struct {
	ID        string    `bson:"_id"`
	Date      time.Time `bson:"date"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	PublicKey string    `bson:"public_key"`
}

type Message struct {
	ID        string    `bson:"_id"`
	ChannelID string    `bson:"channel_id"`
	Date      time.Time `bson:"date"`
	Data      string    `bson:"data"`
	SignedBy  string    `bson:"signed_by"`
}
