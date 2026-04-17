package router

import (
	"acl-challenge/internal/api/handler"

	"github.com/gin-gonic/gin"
)

// Dependencies groups handler dependencies for router initialization.
// It is intentionally empty for the stub phase and can be expanded later.
type Dependencies struct{}

func NewRouter(_ Dependencies) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")

	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)

	protected := v1.Group("/")
	protected.Use(AuthMiddleware())

	notifications := protected.Group("/notifications")

	notifications.GET("/", handler.ListNotifications)
	notifications.POST("/", handler.CreateNotification)
	notifications.GET("/:id", handler.GetNotification)
	notifications.PUT("/:id", handler.UpdateNotification)
	notifications.DELETE("/:id", handler.DeleteNotification)

	users := protected.Group("/users")

	users.PUT("/:id", handler.UpdateUser)
	users.DELETE("/:id", handler.DeleteUser)

	return r
}

// AuthMiddleware is a temporary authentication middleware stub.
// JWT validation will be added in a later milestone.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
