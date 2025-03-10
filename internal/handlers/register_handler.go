package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
)

func RegisterHandler(r repository.UserRepository, tm auth.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request models.RegisterRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		userID, err := r.Register(ctx, request.Login, request.Password)
		if err != nil {
			if errors.Is(err, repository.ErrUserConfict) {
				ctx.AbortWithError(http.StatusConflict, err)
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
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
