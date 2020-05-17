package ws

import (
	"encoding/json"
	"slicerapi/internal/util"
)

type resAddMessage struct {
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Failures []string `json:"failures"`
}

func handleAddMessage(c *client, message wsMessage) {
	// TODO: Use a struct and make better errors.
	content, ok := message.Data["content"]
	iRecipients, ok2 := message.Data["content"]
	if !ok {
		c.send <- []byte("data.content (marshalled & encrypted JSON string) is required")
		return
	} else if !ok2 {
		c.send <- []byte("data.recipients (array of usernames) is required")
		return
	}

	marshalled, err := json.Marshal(wsMessage{
		Method: evtAddMessage,
		Data: map[string]interface{}{
			"signed_by": c.username,
			"content":   content.(string),
		},
	})
	util.Chk(err, true)
	if err != nil {
		c.send <- []byte("Internal server error while marshalling JSON")
		return
	}

	response := resAddMessage{}
	recipients := iRecipients.([]string)
	// TODO: Implement channel broadcasting.
	for _, v := range recipients {
		if c.controller.clients[v] == nil {
			response.Failures = append(response.Failures, v)
		}

		go func(client string) {
			c.controller.clients[client].send <- marshalled
		}(v)
	}

	if len(response.Failures) == len(recipients) {
		response.Message = "Sending to all recipients failed; message still created"
		response.Code = 201
	}

	marshalledResponse, err := json.Marshal(response)
	if err != nil {
		c.send <- []byte("Internal server error while marshalling JSON")
		return
	}
	c.send <- marshalledResponse
}
