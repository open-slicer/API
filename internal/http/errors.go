package http

import (
	"github.com/gin-gonic/gin"
	"slicerapi/internal/logger"
)

// chk is an alternative to util.Chk that aborts with a status code.
func chk(status int, err error, c *gin.Context) {
	if err != nil {
		// Log the error if it's a server error.
		if status >= 500 {
			logger.L.Errorln(err)
		}
		c.JSON(status, gin.H{
			"code": status,
			"message": err.Error(),
		})
	}
}
