package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"slicerapi/internal/config"
	"slicerapi/internal/http/ws"
	"slicerapi/internal/util"
)

// Start starts the HTTP server.
func Start() {
	r := gin.New()
	register(r)
	util.Chk(r.Run(config.C.HTTP.Address))
}

// register registers all routes and middleware.
func register(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowCredentials: true,
	}))

	authMiddlewareFunc := authMiddleware.MiddlewareFunc()

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister)
			auth.POST("/login", authMiddleware.LoginHandler)
			auth.GET("/refresh", authMiddleware.RefreshHandler)
		}

		user := v1.Group("/user")
		user.Use(authMiddlewareFunc)
		{
			user.GET("/:user", handleGetUser)
		}

		channel := v1.Group("/channel")
		channel.Use(authMiddlewareFunc)
		{
			channel.POST("", handleAddChannel)

			specific := channel.Group("/:channel")
			{
				specific.GET("", handleGetChannel)

				specific.GET("/message", handleGetMessage)
				specific.POST("/message", handleAddMessage)

				specific.POST("/join", handleInviteJoin)
				specific.PUT("/user", handleInviteAdd)
			}
		}

		websocket := v1.Group("/ws")
		websocket.Use(authMiddlewareFunc)
		{
			ws.NewController(true)
			go ws.C.Run()

			websocket.GET("", func(c *gin.Context) {
				ws.Handle(c)
			})
		}
	}
}
