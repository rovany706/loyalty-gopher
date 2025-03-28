package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/middleware"
)

func GetUserIDFromContext(ctx *gin.Context) (int, bool) {
	userIDStr, exists := ctx.Get(middleware.UserIDContextKey)
	if !exists {
		return -1, false
	}

	userID, ok := userIDStr.(int)
	if !ok {
		return -1, false
	}

	return userID, true
}
