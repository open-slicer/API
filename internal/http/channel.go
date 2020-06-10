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
	Name     string          `json:"name"`
	Users    map[string]bool `json:"users"`
	Children []string        `json:"children"`
}

type resAddChannel struct {
	statusMessage
	Data db.Channel `json:"data"`
}

type resGetChannel struct {
	statusMessage
	Data []db.Channel `json:"data"`
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

	createdBy := jwt.ExtractClaims(c)["id"].(string)
	id := uuid.New().String()

	channelDoc := db.Channel{
		ID:       id,
		Name:     body.Name,
		Date:     time.Now(),
		Pending:  body.Users,
		Users:    map[string]bool{createdBy: true},
		Children: body.Children,
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").InsertOne(
		ctx,
		channelDoc,
	); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	go func() {
		if createdUser, ok := ws.C.Clients[createdBy]; ok {
			createMarshalled, _ := json.Marshal(ws.Message{
				Method: ws.EvtAddChannel,
				Data:   channelDoc,
			})

			for _, createdClient := range createdUser {
				createdClient.Send <- createMarshalled
			}
		}

		marshalled, _ := json.Marshal(ws.Message{
			Method: ws.EvtAddInvite,
			Data:   channelDoc,
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

		_, _ = db.Mongo.Database(config.C.MongoDB.Name).Collection("users").UpdateOne(
			ctx,
			bson.M{
				"_id": createdBy,
			},
			bson.D{{
				"$push",
				bson.D{{
					"channels",
					channelDoc.ID,
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

func handleGetChannel(c *gin.Context) {
	var channels []db.Channel

	toGet := []string{c.Param("channel")}
	if q := c.Query("for"); q == jwt.ExtractClaims(c)["id"].(string) {
		var user db.User

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

		toGet = append(toGet, user.Channels...)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	cur, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").Find(
		ctx,
		bson.M{
			"_id": bson.D{{
				"$in", toGet,
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
