package router

import (
	"acl-challenge/internal/api/handler"
	"acl-challenge/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler         *handler.AuthHandler
	UserHandler         *handler.UserHandler
	NotificationHandler *handler.NotificationHandler
}

func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")

	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.POST("/logout", deps.AuthHandler.Logout)

	protected := v1.Group("/")
	protected.Use(middleware.ValidateToken(middleware.JWTSecret()))

	notifications := protected.Group("/notifications")

	notifications.GET("/", deps.NotificationHandler.ListNotifications)
	notifications.POST("/", deps.NotificationHandler.CreateNotification)
	notifications.GET("/:id", deps.NotificationHandler.GetNotification)
	notifications.PUT("/:id", deps.NotificationHandler.UpdateNotification)
	notifications.DELETE("/:id", deps.NotificationHandler.DeleteNotification)

	users := protected.Group("/users")

	users.PUT("/:id", deps.UserHandler.UpdateUser)
	users.DELETE("/:id", deps.UserHandler.DeleteUser)

	return r
}
