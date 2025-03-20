package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
)

type AuthHandlers struct {
	userRepository repository.UserRepository
	tokenManager   auth.TokenManager
}

func NewAuthHandlers(r repository.UserRepository, tm auth.TokenManager) *AuthHandlers {
	return &AuthHandlers{
		userRepository: r,
		tokenManager:   tm,
	}
}

func (ah *AuthHandlers) RegisterHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request models.RegisterRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		userID, err := ah.userRepository.Register(ctx, request.Login, request.Password)
		if err != nil {
			if errors.Is(err, repository.ErrUserConfict) {
				ctx.AbortWithError(http.StatusConflict, err)
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		token, err := ah.tokenManager.CreateToken(userID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Header("Authorization", "Bearer "+token)
		ctx.Status(http.StatusOK)
	}
}

func (ah *AuthHandlers) LoginHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request models.LoginRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		userID, err := ah.userRepository.Login(ctx, request.Login, request.Password)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if userID == repository.UnauthorizedUserID {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := ah.tokenManager.CreateToken(userID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Header("Authorization", "Bearer "+token)
		ctx.Status(http.StatusOK)
	}
}
