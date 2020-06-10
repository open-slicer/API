package http

import (
	"context"
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type reqRegister struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	PublicKey string `json:"public_key"`
}

type resGetUser struct {
	statusMessage
	Data db.User `json:"data"`
}

func handleRegister(c *gin.Context) {
	// TODO: nil checks.
	req := reqRegister{}
	err := c.ShouldBind(&req)
	chk(http.StatusBadRequest, err, c)
	if err != nil {
		return
	}

	var user db.User
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(
		ctx,
		bson.M{
			"username": req.Username,
		},
	).Decode(&user); err == nil {
		chk(http.StatusConflict, errors.New("user already exists: "+user.ID), c)
		return
	} else if err != mongo.ErrNoDocuments {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	if len(req.Password) < 10 {
		chk(http.StatusBadRequest, errors.New("password too short; must be at least 10 characters"), c)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Second*2)
	if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").InsertOne(
		ctx,
		db.User{
			ID:        uuid.New().String(),
			Date:      time.Now(),
			Username:  req.Username,
			Password:  string(hash),
			PublicKey: req.PublicKey,
		},
	); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	code := http.StatusCreated
	c.JSON(code, statusMessage{
		Code:    code,
		Message: "Registered; login required.",
	})
}

func handleGetUser(c *gin.Context) {
	var user db.User
	userID := c.Param("user")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(
		ctx,
		bson.M{
			"_id": userID,
		},
	).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusNotFound, err, c)
			return
		}

		chk(http.StatusInternalServerError, err, c)
		return
	}

	reqID := jwt.ExtractClaims(c)["id"].(string)
	if userID != reqID {
		user.Channels = nil
	}
	user.Password = ""

	code := http.StatusOK
	c.JSON(code, resGetUser{
		statusMessage: statusMessage{
			Code:    code,
			Message: "User fetched.",
		},
		Data: user,
	})
}
