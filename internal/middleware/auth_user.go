package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
)

const UserIDContextKey = "user_id"

type authHeader struct {
	Token string `header:"Authorization"`
}

func AuthUser(tm auth.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := authHeader{}

		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		tokenHeader := strings.Split(h.Token, "Bearer ")
		if len(tokenHeader) < 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := tm.GetClaimsFromToken(tokenHeader[1])
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		ctx.Set(UserIDContextKey, claims.UserID)
		ctx.Next()
	}
}
