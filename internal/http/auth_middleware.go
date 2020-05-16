package http

import (
	"crypto/rand"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"slicerapi/internal/db"
	"slicerapi/internal/util"
)

type requestLogin struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type user struct {
	Username string
}

var authMiddleware *jwt.GinJWTMiddleware

func init() {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	util.Chk(err)

	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "keiio",
		Key:         key,
		IdentityKey: "username",
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*user); ok {
				return jwt.MapClaims{
					"username": v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &user{
				Username: claims["username"].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var req requestLogin
			if err := c.ShouldBind(&req); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			password, err := db.Redis.Get("user:" + req.Username).Result()
			if err == db.Nil {
				return nil, jwt.ErrFailedAuthentication
			}
			chk(http.StatusInternalServerError, err, c)
			if err != nil {
				return nil, err
			}

			err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password))
			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user{
				Username: req.Username,
			}, nil
		},
	})
	util.Chk(err)
}
