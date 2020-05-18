package http

import (
	"errors"
	"github.com/gocql/gocql"
	"net/http"
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

func handleRegister(c *gin.Context) {
	req := reqRegister{}
	chk(http.StatusBadRequest, c.ShouldBind(&req), c)

	if len(req.Password) < 10 {
		chk(http.StatusBadRequest, errors.New("password too short; must be at least 10 characters"), c)
		return
	}

	// TODO: Don't actually scan for an ID.
	var id string
	if err := db.Cassandra.Query("SELECT id FROM user WHERE username = ? LIMIT 1", req.Username).Scan(&id); err == nil {
		chk(http.StatusConflict, errors.New("user already exists: "+id), c)
		return
	} else if err != gocql.ErrNotFound {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	chk(http.StatusInternalServerError, err, c)
	if err != nil {
		return
	}

	if err := db.Cassandra.Query(
		"INSERT INTO user (id, date, username, password) VALUES (?, ?, ?, ?)",
		gocql.TimeUUID(),
		time.Now(),
		req.Username,
		string(hash),
	).Exec(); err != nil {
		chk(http.StatusInternalServerError, err, c)
		return
	}

	code := http.StatusCreated
	c.JSON(code, statusMessage{
		Code:    code,
		Message: "Registered; login required.",
	})
}
