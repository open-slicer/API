package http

import (
	"context"
	"encoding/json"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"time"
)

type reqAddChannel struct {
	Name   string          `json:"name"`
	Users  map[string]bool `json:"users"`
	Parent string          `json:"parent"`
}

type resAddChannel struct {
	statusMessage
	Data     db.Channel `json:"data"`
	Failures []string   `json:"failures"`
}

type resGetChannel struct {
	statusMessage
	Data db.Channel `json:"data"`
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
	id := uuid.New().String()

	channelDoc := db.Channel{
		ID:      id,
		Name:    body.Name,
		Date:    time.Now(),
		Pending: body.Users,
		Users:   map[string]bool{createdBy: true},
		Parent:  body.Parent,
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").InsertOne(
		ctx,
		channelDoc,
	); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	data := map[string]interface{}{
		"created_by": createdBy,
		"id":         id,
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
		Data: channelDoc,
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

	for i := range body.Users {
		if ws.C.Clients[i] == nil {
			response.Failures = append(response.Failures, i)
			continue
		}

		go func(user string) {
			for _, client := range ws.C.Clients[user] {
				client.Send <- marshalled
			}
		}(i)
	}

	c.JSON(response.Code, response)
}

func handleGetChannel(c *gin.Context) {
	var channel db.Channel

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").FindOne(
		ctx,
		bson.M{
			"_id": c.Param("channel"),
		},
	).Decode(&channel); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusNotFound, err, c)
			return
		}

		chk(http.StatusInternalServerError, err, c)
		return
	}

	code := http.StatusOK
	c.JSON(code, resGetChannel{
		statusMessage: statusMessage{
			Code:    code,
			Message: "Channel fetched.",
		},
		Data: channel,
	})
}
