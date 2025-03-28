package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/helpers"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/shopspring/decimal"
)

type PointsHandlers struct {
	pointsRepository repository.PointsRepository
}

func NewPointsHandlers(pr repository.PointsRepository) *PointsHandlers {
	return &PointsHandlers{
		pointsRepository: pr,
	}
}

func (ph *PointsHandlers) UserBalanceHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := helpers.GetUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		balance, err := ph.pointsRepository.GetUserBalance(ctx, userID)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		withdrawalHistory, err := ph.pointsRepository.GetUserWithdrawalHistory(ctx, userID)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		sum := decimal.Zero
		for _, entry := range withdrawalHistory {
			sum = sum.Add(entry.WithdrawSum)
		}

		response := models.GetUserBalanceResponse{
			Current:   balance,
			Withdrawn: sum,
		}

		ctx.JSON(http.StatusOK, response)
	}
}

func (ph *PointsHandlers) WithdrawPointsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := helpers.GetUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var request models.WithdrawUserPointsRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ok, err := helpers.LuhnCheck(request.OrderNum)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !ok {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		err = ph.pointsRepository.WithdrawPoints(ctx, userID, request.OrderNum, request.WithdrawSum)
		if err != nil {
			if errors.Is(err, repository.ErrNotEnoughPoints) {
				ctx.AbortWithStatus(http.StatusPaymentRequired)
				return
			}

			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

func (ph *PointsHandlers) GetUserWithdrawalHistory() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := helpers.GetUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		withdrawalHistory, err := ph.pointsRepository.GetUserWithdrawalHistory(ctx, userID)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if len(withdrawalHistory) == 0 {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.JSON(http.StatusOK, withdrawalHistory)
	}
}
