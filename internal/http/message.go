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
			"signed_by": jwt.ExtractClaims(c)["id"],
			"data":      body.Data,
		},
	})
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	code := http.StatusNotFound
	channel, ok := ws.C.Channels[c.Param("channel")]
	if !ok {
		c.JSON(code, statusMessage{
			Message: "Invalid channel ID.",
			Code:    code,
		})
		return
	}

	channel.Send <- marshalled

	code = http.StatusCreated
	c.JSON(code, statusMessage{
		Message: "Message created.",
		Code:    code,
	})
}
