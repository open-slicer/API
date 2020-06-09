// ! This doesn't actually work. It's simply here to show how it'd be done. See README.md.

package http

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"net/http"
	"slicerapi/internal/db"
)

func handleInviteJoin(c *gin.Context) {
	userID := jwt.ExtractClaims(c)["id"].(string)
	chID := c.Param("id")

	var pending []string
	var users []string
	if err := db.Cassandra.Query(
		"SELECT pending, users FROM channel WHERE id = ? LIMIT 1",
		chID,
	).Scan(&pending, &users); err != nil {
		if err == gocql.ErrNotFound {
			chk(http.StatusUnauthorized, err, c)
			return
		}
		chk(http.StatusInternalServerError, err, c)
		return
	}

	// TODO: Improve this. (for the 50th time)
	authorized := false
	for i, v := range pending {
		if v == userID {
			authorized = true

			length := len(pending)
			pending[length-1], pending[i] = pending[i], pending[length-1]
			pending = pending[:length-1]
			break
		}
	}

	stat := http.StatusOK
	if authorized {
		users = append(users, userID)
		err := db.Cassandra.Query("UPDATE channel SET pending = ?, users = ? WHERE id = ?", pending, users, chID).Exec()
		chk(500, err, c)
		if err != nil {
			return
		}

		c.JSON(stat, statusMessage{
			Code:    stat,
			Message: "Invite accepted.",
		})
	}
	stat = http.StatusForbidden
	c.JSON(stat, statusMessage{
		Code:    stat,
		Message: "User isn't in the pending list.",
	})
}
