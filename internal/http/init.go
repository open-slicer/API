package http

import (
	"github.com/gin-gonic/gin"
	"slicerapi/internal/http/ws"
	"slicerapi/internal/util"
)

func Start() {
	r := gin.New()
	register(r)
	util.Chk(r.Run(util.Config.HTTP.Address))
}

// register registers all routes and middleware.
func register(r *gin.Engine) {
	authMiddlewareFunc := authMiddleware.MiddlewareFunc()

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister)
			auth.POST("/login", authMiddleware.LoginHandler)
			auth.GET("/refresh", authMiddleware.RefreshHandler)
		}

		messages := v1.Group("/messages")
		messages.Use(authMiddlewareFunc)
		{
		}

		websocket := v1.Group("/ws")
		websocket.Use(authMiddlewareFunc)
		{
			controller := ws.NewController()
			go controller.Run()

			websocket.GET("", func(c *gin.Context) {
				ws.Handle(controller, c)
			})
		}
	}
}
