package http

import (
	"slicerapi/internal/logger"

	"github.com/gin-gonic/gin"
)

// statusMessage is a generic message sent to clients to specify the current status.
type statusMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// chk is an alternative to util.Chk that aborts with a status code.
func chk(status int, err error, c *gin.Context) {
	if err != nil {
		// Log the error if it's a server error.
		if status >= 500 {
			logger.L.Errorln(err)
		}
		c.JSON(status, statusMessage{
			Code:    status,
			Message: err.Error(),
		})
	}
}
