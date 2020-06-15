package http

import (
	"context"
	"encoding/json"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type reqAddChannel struct {
	Name     string          `json:"name"`
	Users    map[string]bool `json:"users"`
	Children []string        `json:"children"`
	Parent   string          `json:"parent"`
}

type resAddChannel struct {
	statusMessage
	Data db.Channel `json:"data"`
}

type resGetChannel struct {
	statusMessage
	Data []db.Channel `json:"data"`
}

// handleAddChannel handles requests asking for channels to be created.
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
		ID:       id,
		Name:     body.Name,
		Date:     time.Now(),
		Pending:  body.Users,
		Users:    map[string]bool{createdBy: true},
		Children: body.Children,
		Parent:   body.Parent,
		Owner:    createdBy,
	}

	// Insert the channel.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").InsertOne(
		ctx,
		channelDoc,
	); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	go func() {
		// Send the EVT_ADD_CHANNEL event to createdBy over ws.
		if createdUser, ok := ws.C.Clients[createdBy]; ok {
			createMarshalled, _ := json.Marshal(ws.ChannelMessage{
				Message: ws.Message{Method: ws.EvtAddChannel},
				Data:    channelDoc,
			})

			for _, createdClient := range createdUser {
				createdClient.Send <- createMarshalled
			}
		}

		// Send EVT_ADD_INVITE to every invited user over ws.
		marshalled, _ := json.Marshal(ws.ChannelMessage{
			Message: ws.Message{Method: ws.EvtAddInvite},
			Data:    channelDoc,
		})

		for i := range body.Users {
			client, ok := ws.C.Clients[i]
			if !ok {
				continue
			}

			go func() {
				for _, conn := range client {
					conn.Send <- marshalled
				}
			}()
		}

		// Push the new channel to the user who created the channel's channels array.
		_, _ = db.Mongo.Database(config.C.MongoDB.Name).Collection("users").UpdateOne(
			ctx,
			bson.M{
				"_id": createdBy,
			},
			bson.D{{
				Key: "$push",
				Value: bson.D{{
					Key:   "channels",
					Value: channelDoc.ID,
				}},
			}},
		)
	}()

	response := resAddChannel{
		statusMessage: statusMessage{
			Message: "Channel created.",
			Code:    http.StatusCreated,
		},
		Data: channelDoc,
	}

	c.JSON(response.Code, response)
}

// handleGetChannel handles requests asking for channel info.
func handleGetChannel(c *gin.Context) {
	var channels []db.Channel

	toGet := []string{c.Param("channel")}
	// If there's a `for` query param:
	if q := c.Query("for"); q == jwt.ExtractClaims(c)["id"].(string) {
		var user db.User

		// Find the `for` user.
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(
			ctx,
			bson.M{
				"_id": q,
			},
		).Decode(&user); err != nil {
			if err == mongo.ErrNoDocuments {
				chk(http.StatusNotFound, err, c)
				return
			}

			chk(http.StatusInternalServerError, err, c)
			return
		}

		// Append to toGet with the user's channels.
		toGet = append(toGet, user.Channels...)
	}

	// Find each channel.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	cur, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").Find(
		ctx,
		bson.M{
			"_id": bson.D{{
				Key:   "$in",
				Value: toGet,
			}},
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

	// Extract the channels.
	ctx, _ = context.WithTimeout(context.Background(), time.Second*2)
	err = cur.All(ctx, &channels)
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	code := http.StatusOK
	c.JSON(code, resGetChannel{
		statusMessage: statusMessage{
			Code:    code,
			Message: "Channels fetched.",
		},
		Data: channels,
	})
}
