package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"slicerapi/internal/db"
)

func handleRegister(c *gin.Context) {
	req := requestLogin{}
	chk(http.StatusBadRequest, c.ShouldBindJSON(&req), c)

	_, err := db.Client.Get("user:" + req.Username).Result()
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

	go db.Client.Set("user:"+req.Username, hash, 0)

	code := http.StatusCreated
	c.JSON(code, gin.H{
		"code": code,
		"message": "registered; login required",
		"username": req.Username,
	})
}
