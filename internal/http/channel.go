package http

import (
	"encoding/json"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"time"
)

type reqAddChannel struct {
	Name   string   `json:"name"`
	Users  []string `json:"users"`
	Parent string   `json:"parent"`
}

type channelData struct {
	ID string `json:"id"`
}

type resAddChannel struct {
	statusMessage
	Data     channelData `json:"data"`
	Failures []string    `json:"failures"`
}

func handleAddChannel(c *gin.Context) {
	body := reqAddChannel{}
	err := c.ShouldBindJSON(&body)
	chk(http.StatusBadRequest, err, c)
	if err != nil {
		return
	}

	if body.Name == "" {
		body.Name = "New Channel"
	}
	if body.Parent == "" {
		body.Parent = "slicer_origin"
	}

	createdBy := jwt.ExtractClaims(c)["id"].(string)
	id := gocql.TimeUUID()
	if err := db.Cassandra.Query(
		"INSERT INTO channel (id, name, date, pending, users, parent) VALUES (?, ?, ?, ?, ?, ?)",
		id,
		body.Name,
		time.Now(),
		body.Users,
		[]string{createdBy},
		body.Parent,
	).Exec(); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	idString := id.String()
	data := map[string]interface{}{
		"created_by": createdBy,
		"id":         idString,
	}
	marshalled, err := json.Marshal(ws.Message{
		Method: ws.EvtAddInvite,
		Data:   data,
	})
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	response := resAddChannel{
		statusMessage: statusMessage{
			Message: "Channel created.",
			Code:    http.StatusCreated,
		},
		Data: channelData{
			ID: idString,
		},
	}

	var createMarshalled []byte
	if createdUser := ws.C.Clients[createdBy]; createdUser != nil {
		var err error
		createMarshalled, err = json.Marshal(ws.Message{
			Method: ws.EvtAddChannel,
			Data:   data,
		})

		chk(http.StatusInternalServerError, err, c)
		if err != nil {
			return
		}

		for _, createdClient := range createdUser {
			createdClient.Send <- createMarshalled
		}
	} else {
		response.Failures = append(response.Failures, createdBy)
	}

	for _, v := range body.Users {
		if ws.C.Clients[v] == nil {
			response.Failures = append(response.Failures, v)
			continue
		}

		go func(user string) {
			for _, client := range ws.C.Clients[user] {
				client.Send <- marshalled
			}
		}(v)
	}

	c.JSON(response.Code, response)
}
