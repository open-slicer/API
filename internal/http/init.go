package http

import (
	"github.com/gin-gonic/gin"
	"slicerapi/internal/util"
)

func Start() {
	r := gin.New()
	register(r)
	util.Chk(r.Run(util.Config.HTTP.Address))
}

// register registers all routes and middleware.
func register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister)
			auth.POST("/login", authMiddleware.LoginHandler)
			auth.GET("/refresh", authMiddleware.RefreshHandler)
		}

		ws := v1.Group("/ws")
		ws.Use(authMiddleware.MiddlewareFunc())
		{
			ws.GET("", func(c *gin.Context) {
				handleWS(c.Writer, c.Request)
			})
		}
	}
}
