package http

import (
	"encoding/json"
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"slicerapi/internal/http/ws"
)

// TODO: Actually store messages in the Cassandra cluster.

type reqAddMessage struct {
	Data       string   `json:"data"`
	Recipients []string `json:"recipients"`
}

type resAddMessage struct {
	statusMessage
	Failures []string `json:"failures"`
}

func handleAddMessage(c *gin.Context) {
	body := reqAddMessage{}
	err := c.ShouldBindJSON(&body)
	chk(http.StatusBadRequest, err, c)
	if err != nil {
		return
	}

	if body.Data == "" {
		chk(http.StatusBadRequest, errors.New("`data` is required but is missing"), c)
		return
	}

	marshalled, err := json.Marshal(ws.Message{
		Method: ws.EvtAddMessage,
		Data: map[string]interface{}{
			"signed_by": jwt.ExtractClaims(c)["username"],
			"data":      body.Data,
		},
	})
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	response := resAddMessage{}
	// TODO: Implement channel broadcasting.
	for _, v := range body.Recipients {
		if ws.C.Clients[v] == nil {
			response.Failures = append(response.Failures, v)
			continue
		}

		go func(client string) {
			ws.C.Clients[client].Send <- marshalled
		}(v)
	}

	code := http.StatusCreated
	if len(response.Failures) == len(body.Recipients) {
		response.Message = "Sending to all recipients failed; message still created"
		response.Code = code
	}

	c.JSON(code, response)
}
