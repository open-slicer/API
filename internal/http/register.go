package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"slicerapi/internal/db"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type userModel struct {
	Username  string `json:"username"`
	Hash      []byte `json:"hash"`
	PublicKey string `json:"public_key"`
}

type requestRegister struct {
	Username  string `form:"username" json:"username"`
	Password  string `form:"password" json:"password"`
	PublicKey string `json:"public_key"`
}

func handleRegister(c *gin.Context) {
	req := requestRegister{}
	chk(http.StatusBadRequest, c.ShouldBind(&req), c)

	_, err := db.Redis.Get("user:" + req.Username).Result()
	if err == nil {
		chk(http.StatusConflict, errors.New("user already exists"), c)
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

	user := userModel{
		Username:  req.Username,
		Hash:      hash,
		PublicKey: req.PublicKey,
	}
	marshalled, err := json.Marshal(user)
	go db.Redis.Set("user:"+req.Username, marshalled, 0)

	code := http.StatusCreated
	c.JSON(code, gin.H{
		"code":    code,
		"message": "registered; login required",
	})
}
