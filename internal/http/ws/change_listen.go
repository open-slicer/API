package ws

import (
	"context"
	"encoding/json"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type changeListenMessage struct {
	Message
	Data []string `json:"data"`
}

func handleChangeListen(c *Client, msg changeListenMessage) {
	if len(msg.Data) < 1 {
		var user db.User
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)

		if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(ctx, bson.M{
			"_id": c.ID,
		}).Decode(&user); err != nil {
			marshalled, _ := json.Marshal(ErrMessage{
				Message: Message{Method: errDB},
				Data:    err.Error(),
			})

			c.Send <- marshalled
			return
		}

		msg.Data = user.Channels
	}

	for _, v := range msg.Data {
		channel, ok := C.Channels[v]
		if !ok {
			var err error
			channel, err = NewChannel(v)

			if err != nil {
				marshalled, _ := json.Marshal(ErrMessage{
					Message: Message{Method: errInvalidArgument},
					Data:    "channel_id",
				})

				c.Send <- marshalled
				return
			}

			go channel.Listen()
		}

		if _, ok := channel.Clients[v]; ok {
			channel.unregister <- c
		} else {
			channel.register <- c
		}
	}

	marshalled, _ := json.Marshal(Message{
		Method: evtChangeListen,
	})

	c.Send <- marshalled
}
