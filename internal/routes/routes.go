package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/handlers"
	"github.com/rovany706/loyalty-gopher/internal/middleware"
)

func RegisterAuthHandlers(r *gin.Engine, authHandlers *handlers.AuthHandlers) {
	authGroup := r.Group("/api/user")
	{
		authGroup.POST("/register", authHandlers.RegisterHandler())
		authGroup.POST("/login", authHandlers.LoginHandler())
	}
}

func RegisterOrderHandlers(r *gin.Engine, orderHandlers *handlers.OrderHandlers, tm auth.TokenManager) {
	orderGroup := r.Group("/api/user")
	{
		orderGroup.Use(middleware.AuthUser(tm))
		orderGroup.POST("/orders", orderHandlers.PostNewOrderHandler())
		orderGroup.GET("/orders", orderHandlers.GetUserOrdersHandler())
	}
}
