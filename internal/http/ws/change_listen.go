package ws

import (
	"encoding/json"
	"slicerapi/internal/util"
)

func handleChangeListen(c *Client, msg Message) {
	chID, ok := msg.Data["channel_id"]
	if !ok {
		marshalled, err := json.Marshal(Message{
			Method: errMissingArgument,
			Data: map[string]interface{}{
				"arg": "channel_id",
			},
		})
		if err != nil {
			util.Chk(err, true)
			c.Send <- []byte(errJSON)
		}

		c.Send <- marshalled
		return
	}

	strID := chID.(string)
	channel, ok := C.Channels[strID]
	if !ok {
		var err error
		channel, err = NewChannel(strID)
		C.Channels[strID] = channel

		if err != nil {
			marshalled, err := json.Marshal(Message{
				Method: errInvalidArgument,
				Data: map[string]interface{}{
					"arg": "channel_id",
				},
			})
			if err != nil {
				util.Chk(err, true)
				c.Send <- []byte(errJSON)
			}

			c.Send <- marshalled
			return
		}

		go channel.Listen()
	}

	if _, ok := channel.Clients[strID]; ok {
		channel.unregister <- c
	} else {
		channel.register <- c
	}

	marshalled, err := json.Marshal(Message{
		Method: evtChangeListen,
	})
	if err != nil {
		util.Chk(err, true)
		c.Send <- []byte(errJSON)
	}

	c.Send <- marshalled
}
