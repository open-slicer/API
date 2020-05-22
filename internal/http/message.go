package http

import (
	"encoding/json"
	"errors"
	"github.com/gocql/gocql"
	"net/http"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"slicerapi/internal/util"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type reqAddMessage struct {
	Data string `json:"data"`
}

type messageData struct {
	ID string `json:"id"`
}

type resAddMessage struct {
	statusMessage
	Data messageData `json:"data"`
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

	signedID := jwt.ExtractClaims(c)["id"]
	marshalled, err := json.Marshal(ws.Message{
		Method: ws.EvtAddMessage,
		Data: map[string]interface{}{
			"signed_by": signedID,
			"data":      body.Data,
		},
	})
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	code := http.StatusNotFound
	chID := c.Param("channel")
	channel, ok := ws.C.Channels[chID]
	if !ok {
		channel, err = ws.NewChannel(chID)

		if err != nil {
			util.Chk(err, true)
			c.JSON(code, statusMessage{
				Message: "Invalid channel ID.",
				Code:    code,
			})
			return
		}

		go channel.Listen()
	}

	id := gocql.TimeUUID()
	go func() {
		parsedSignedUUID, _ := gocql.ParseUUID(signedID.(string))
		go db.Cassandra.Query(
			"INSERT INTO message (id, data, date, signed_by) VALUES (?, ?, ?, ?)",
			id,
			body.Data,
			time.Now(),
			parsedSignedUUID,
		).Exec()

		channel.Send <- marshalled
	}()

	code = http.StatusCreated
	c.JSON(code, resAddMessage{
		statusMessage: statusMessage{
			Message: "Message created.",
			Code:    code,
		},
		Data: messageData{ID: id.String()},
	})
}
