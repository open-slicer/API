package db

import (
	"time"
)

type Channel struct {
	ID      string          `bson:"_id" json:"id,omitempty"`
	Name    string          `bson:"name" json:"name,omitempty"`
	Date    time.Time       `bson:"date" json:"date,omitempty"`
	Pending map[string]bool `bson:"pending" json:"pending,omitempty"`
	Users   map[string]bool `bson:"users" json:"users,omitempty"`
	Parent  string          `bson:"parent" json:"parent,omitempty"`
}

type User struct {
	ID        string    `bson:"_id" json:"id,omitempty"`
	Date      time.Time `bson:"date" json:"date,omitempty"`
	Username  string    `bson:"username" json:"username,omitempty"`
	Password  string    `bson:"password" json:"password,omitempty"`
	PublicKey string    `bson:"public_key" json:"public_key,omitempty"`
}

type Message struct {
	ID        string    `bson:"_id" json:"id,omitempty"`
	ChannelID string    `bson:"channel_id" json:"channel_id,omitempty"`
	Date      time.Time `bson:"date" json:"date,omitempty"`
	Data      string    `bson:"data" json:"data,omitempty"`
	SignedBy  string    `bson:"signed_by" json:"signed_by,omitempty"`
}
