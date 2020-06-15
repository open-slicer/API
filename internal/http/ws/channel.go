package ws

import (
	"context"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Channel is a realm in which users can perform actions, like sending messages.
type Channel struct {
	Clients         map[string][]*Client
	Send            chan []byte
	possibleClients map[string]bool
	unregister      chan *Client
	register        chan *Client
}

// NewChannel instantiates a channel.
func NewChannel(chID string) (*Channel, error) {
	var chDoc db.Channel

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	// Find the channel with the UUID of chID.
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").FindOne(ctx, bson.M{
		"_id": chID,
	}).Decode(&chDoc); err != nil {
		return nil, err
	}

	channel := &Channel{
		Clients:         make(map[string][]*Client),
		Send:            make(chan []byte),
		possibleClients: chDoc.Users,
		register:        make(chan *Client),
		unregister:      make(chan *Client),
	}

	C.Channels[chID] = channel
	return channel, nil
}

// Listen listens for channel inputs. This blocks.
func (ch *Channel) Listen() {
	for {
		select {
		case client := <-ch.register:
			if _, ok := ch.possibleClients[client.ID]; ok {
				ch.Clients[client.ID] = append(ch.Clients[client.ID], client)
			}
		case client := <-ch.unregister:
			if c, ok := ch.Clients[client.ID]; ok {
				// Delete the client from all stores.
				length := len(c)
				c[length-1], c[client.index] = c[client.index], c[length-1]
				c = c[:length-1]

				if len(c) <= 0 {
					delete(ch.Clients, client.ID)
				}
			}
		case message := <-ch.Send:
			for _, user := range ch.Clients {
				for _, client := range user {
					select {
					case client.Send <- message:
					default:
						ch.unregister <- client
					}
				}
			}
		}
	}
}
