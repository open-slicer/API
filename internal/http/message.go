package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"slicerapi/internal/util"
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type reqAddMessage struct {
	Data string `json:"data"`
}

type resAddMessage struct {
	statusMessage
	Data db.Message `json:"data"`
}

type resGetMessage struct {
	statusMessage
	Data []db.Message `json:"data"`
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

	var chDoc db.Channel

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").FindOne(
		ctx,
		bson.M{
			"_id": chID,
		},
	).Decode(&chDoc); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(code, err, c)
			return
		}

		chk(http.StatusInternalServerError, err, c)
		return
	}

	signedID := jwt.ExtractClaims(c)["id"].(string)
	_, ok := chDoc.Users[signedID]
	if !ok {
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

	newMsg := db.Message{
		ID:        uuid.New().String(),
		Date:      time.Now(),
		Data:      body.Data,
		SignedBy:  signedID,
		ChannelID: chID,
	}
	go func() {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		go db.Mongo.Database(config.C.MongoDB.Name).Collection("messages").InsertOne(
			ctx,
			newMsg,
		)

		marshalled, _ := json.Marshal(ws.Message{
			Method: ws.EvtAddMessage,
			Data: newMsg,
		})
		channel.Send <- marshalled
	}()

	code = http.StatusCreated
	c.JSON(code, resAddMessage{
		statusMessage: statusMessage{
			Message: "Message created.",
			Code:    code,
		},
		Data: newMsg,
	})
}

func handleGetMessage(c *gin.Context) {
	var limit int64 = 50
	if limitStr := c.Query("limit"); limitStr != "" {
		var err error

		limit, err = strconv.ParseInt(limitStr, 10, 32)
		chk(http.StatusBadRequest, err, c)
		if err != nil {
			return
		}

		if limit > 100 {
			chk(http.StatusBadRequest, errors.New("limit must be <= 100"), c)
			return
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	cur, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("messages").Find(
		ctx,
		bson.M{
			"channel_id": c.Param("channel"),
		},
		&options.FindOptions{
			Limit: &limit,
		},
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusNotFound, err, c)
			return
		}
		chk(http.StatusInternalServerError, err, c)
		return
	}

	var res []db.Message

	ctx, _ = context.WithTimeout(context.Background(), time.Second*2)
	err = cur.All(ctx, &res)
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	stat := http.StatusOK
	c.JSON(stat, resGetMessage{
		statusMessage: statusMessage{
			Code:    stat,
			Message: "Messages fetched.",
		},
		Data: res,
	})
}
