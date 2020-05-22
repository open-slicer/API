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

	code := http.StatusNotFound
	chID := c.Param("channel")

	var usersSlice []string
	if err := db.Cassandra.Query("SELECT users FROM channel WHERE id = ? LIMIT 1", chID).Scan(&usersSlice); err != nil {
		if err == gocql.ErrNotFound {
			chk(code, err, c)
			return
		}

		chk(http.StatusInternalServerError, err, c)
		return
	}

	// TODO: Again, this really, really needs improving. It's simply for testing.
	signedID := jwt.ExtractClaims(c)["id"].(string)
	authorized := false
	for _, v := range usersSlice {
		if v == signedID {
			authorized = true
			break
		}
	}
	if !authorized {
		chk(http.StatusForbidden, errors.New("can't send messages in this channel; not in users set"), c)
		return
	}

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

	id := gocql.TimeUUID()
	go func() {
		go db.Cassandra.Query(
			"INSERT INTO message (id, data, date, signed_by) VALUES (?, ?, ?, ?)",
			id,
			body.Data,
			time.Now(),
			signedID,
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
