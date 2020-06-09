package http

import (
	"context"
	"crypto/rand"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"slicerapi/internal/util"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type requestLogin struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type user struct {
	ID       string
	Username string
}

var authMiddleware *jwt.GinJWTMiddleware

func init() {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	util.Chk(err)

	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "slicer",
		Key:         key,
		IdentityKey: "username",
		// Not sure as to whether or not this should be longer.
		MaxRefresh:  time.Hour * 2,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*user); ok {
				return jwt.MapClaims{
					"id":       v.ID,
					"username": v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &user{
				ID:       claims["id"].(string),
				Username: claims["username"].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var req requestLogin
			if err := c.ShouldBind(&req); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var userDoc db.User

			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(ctx, bson.M{
				"username": req.Username,
			}).Decode(&userDoc); err != nil {
				fmt.Println(err)
				return nil, jwt.ErrFailedAuthentication
			}

			if err = bcrypt.CompareHashAndPassword([]byte(userDoc.Password), []byte(req.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user{
				Username: req.Username,
				ID:       userDoc.ID,
			}, nil
		},
	})
	util.Chk(err)
}
