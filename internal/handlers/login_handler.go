package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
)

func LoginHandler(r repository.UserRepository, tm auth.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request models.LoginRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		userID, err := r.Login(ctx, request.Login, request.Password)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if userID == repository.UnauthorizedUserID {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := tm.CreateToken(userID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Header("Authorization", "Bearer "+token)
		ctx.Status(http.StatusOK)
	}
}
