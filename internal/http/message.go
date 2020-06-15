package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"slicerapi/internal/util"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

// handleAddMessage is the main handler for requests asking for messages to be sent.
// This should be as high-performance as possible; we want to keep message send times as low as possible.
// Currently — on a local network — this takes around 7ms. The overhead is due to querying the DB.
// This also assumes that inserting and sending completed; unsure as to whether or not we should confirm this.
// It's essentially reliant on the status of the channel query.
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

	// Find the channel.
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

	// If the user isn't in the users array, abort.
	signedID := jwt.ExtractClaims(c)["id"].(string)
	_, ok := chDoc.Users[signedID]
	if !ok {
		chk(
			http.StatusForbidden,
			errors.New("can't send messages in this channel; not in users array. use the /api/v1/:channel/join endpoint"),
			c,
		)
		return
	}

	// Create the websocket channel if it exists.
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
		// Insert the new message and send it over ws.
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		go db.Mongo.Database(config.C.MongoDB.Name).Collection("messages").InsertOne(
			ctx,
			newMsg,
		)

		marshalled, _ := json.Marshal(ws.ChatMessage{
			Message: ws.Message{Method: ws.EvtAddMessage},
			Data:    newMsg,
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

// handleGetMessage handles requests asking for 1 or more messages.
func handleGetMessage(c *gin.Context) {
	// The default limit is 50. This can go up to 100.
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

	// Get the messages in the channel with the limit.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	cur, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("messages").Find(
		ctx,
		bson.M{
			"channel_id": c.Param("channel"),
			"date":       bson.M{"$gte": c.Query("from"), "$lte": c.Query("to")},
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

	// Decode the results into res.
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
