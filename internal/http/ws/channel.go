package ws

import (
	"slicerapi/internal/db"
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
	var usersSlice []string
	if err := db.Cassandra.Query("SELECT users FROM channel WHERE id = ? LIMIT 1", chID).Scan(&usersSlice); err != nil {
		return nil, err
	}

	// TODO: This is inefficient. Do something else.
	users := map[string]bool{}
	for _, v := range usersSlice {
		users[v] = true
	}

	channel := &Channel{
		Clients:         make(map[string][]*Client),
		Send:            make(chan []byte),
		possibleClients: users,
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
