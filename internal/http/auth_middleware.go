package http

import (
	"crypto/rand"
	"slicerapi/internal/db"
	"slicerapi/internal/util"

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

			var id string
			var password string
			if err := db.Cassandra.Query(
				"SELECT id, password FROM user WHERE username = ?",
				req.Username,
			).Scan(&id, &password); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user{
				Username: req.Username,
				ID:       id,
			}, nil
		},
	})
	util.Chk(err)
}
