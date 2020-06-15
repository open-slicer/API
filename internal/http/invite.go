package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/http/ws"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type reqInviteAdd struct {
	ID string `json:"id"`
}

// handleInviteJoin handles requests asking to join channels.
// This will only work if the user has previously been added.
func handleInviteJoin(c *gin.Context) {
	userID := jwt.ExtractClaims(c)["id"].(string)
	chID := c.Param("channel")

	var channel db.Channel

	// Find the channel.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").FindOne(
		ctx,
		bson.M{
			"_id": chID,
		},
	).Decode(&channel); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusUnauthorized, err, c)
			return
		}
		chk(http.StatusInternalServerError, err, c)
		return
	}

	stat := http.StatusOK

	_, ok := channel.Pending[userID]
	if ok {
		// If the user is pending (invited), accept them and move their user ID from the pending array to the users array.
		delete(channel.Pending, userID)

		channel.Users[userID] = true
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").UpdateOne(
			ctx,
			bson.M{
				"_id": channel.ID,
			},
			bson.D{{
				Key: "$set",
				Value: bson.D{{
					Key:   "pending",
					Value: channel.Pending,
				}, {
					Key:   "users",
					Value: channel.Users,
				}},
			}},
		); err != nil {
			chk(500, err, c)
			return
		}

		go func() {
			// Also add the channel to the user's channels array.
			_, _ = db.Mongo.Database(config.C.MongoDB.Name).Collection("users").UpdateOne(
				ctx,
				bson.M{
					"_id": userID,
				},
				bson.D{{
					Key: "$push",
					Value: bson.D{{
						Key:   "channels",
						Value: channel.ID,
					}},
				}},
			)
		}()

		c.JSON(stat, statusMessage{
			Code:    stat,
			Message: "Invite accepted.",
		})
		return
	}

	stat = http.StatusForbidden
	c.JSON(stat, statusMessage{
		Code:    stat,
		Message: "User isn't in the pending list.",
	})
}

// handleInviteAdd handles requests asking for invites to be created.
// This allows the handleInviteJoin endpoint to be used.
func handleInviteAdd(c *gin.Context) {
	body := reqInviteAdd{}
	err := c.ShouldBindJSON(&body)
	chk(http.StatusBadRequest, err, c)
	if err != nil {
		return
	}

	chID := c.Param("channel")
	var channel db.Channel

	// Find the channel.
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").FindOne(
		ctx,
		bson.M{
			"_id": chID,
		},
	).Decode(&channel); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusUnauthorized, err, c)
			return
		}
		chk(http.StatusInternalServerError, err, c)
		return
	}

	// If the user is already in the channel's users array, they've already joined.
	if _, ok := channel.Users[body.ID]; ok {
		chk(http.StatusConflict, errors.New("user already joined"), c)
		return
	}

	// Make sure Pending actually exists before editing it.
	if channel.Pending == nil {
		channel.Pending = map[string]bool{}
	}
	// Make the user pending.
	channel.Pending[body.ID] = true

	// Set the new pending value.
	ctx, _ = context.WithTimeout(context.Background(), time.Second*2)
	if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").UpdateOne(
		ctx,
		bson.M{
			"_id": channel.ID,
		},
		bson.D{{
			Key: "$set",
			Value: bson.D{{
				Key:   "pending",
				Value: channel.Pending,
			}},
		}},
	); err != nil {
		chk(500, err, c)
		return
	}

	// Send an EVT_ADD_INVITE event over ws.
	go func() {
		marshalled, _ := json.Marshal(ws.ChannelMessage{
			Message: ws.Message{Method: ws.EvtAddInvite},
			Data:    channel,
		})

		client, ok := ws.C.Clients[body.ID]
		if !ok {
			return
		}

		go func() {
			for _, conn := range client {
				conn.Send <- marshalled
			}
		}()
	}()

	stat := http.StatusCreated
	c.JSON(stat, statusMessage{
		Code:    stat,
		Message: "Invite created for user.",
	})
}
