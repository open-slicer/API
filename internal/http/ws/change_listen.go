package ws

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"time"
)

func handleChangeListen(c *Client, msg Message) {
	chIDs, ok := msg.Data.(map[string][]string)["channel_ids"]
	if !ok {
		var user db.User
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)

		if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(ctx, bson.M{
			"_id": c.ID,
		}).Decode(&user); err != nil {
			marshalled, _ := json.Marshal(Message{
				Method: errDB,
				Data: map[string]string{
					"err": err.Error(),
				},
			})

			c.Send <- marshalled
			return
		}

		chIDs = user.Channels
	}

	for _, v := range chIDs {
		channel, ok := C.Channels[v]
		if !ok {
			var err error
			channel, err = NewChannel(v)

			if err != nil {
				marshalled, _ := json.Marshal(Message{
					Method: errInvalidArgument,
					Data: map[string]interface{}{
						"arg": "channel_id",
					},
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
