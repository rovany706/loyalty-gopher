package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/helpers"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/rovany706/loyalty-gopher/internal/services"
)

type OrderHandlers struct {
	orderRepository repository.OrderRepository
	accrualService  services.AccrualService
}

func NewOrderHandlers(orderRepository repository.OrderRepository, accrualService services.AccrualService) *OrderHandlers {
	return &OrderHandlers{
		orderRepository: orderRepository,
		accrualService:  accrualService,
	}
}

func (oh *OrderHandlers) PostNewOrderHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		orderNum := string(body)

		ok, err := helpers.LuhnCheck(orderNum)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !ok {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		userID, ok := helpers.GetUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		existingOrder, err := oh.orderRepository.GetOrder(ctx, orderNum)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if existingOrder != nil {
			if existingOrder.UserID == userID {
				ctx.Status(http.StatusOK)
				return
			} else {
				ctx.Status(http.StatusConflict)
				return
			}
		}

		err = oh.orderRepository.AddOrder(ctx, userID, orderNum)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = oh.accrualService.QueueStatusUpdate(ctx, orderNum)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusAccepted)
	}
}

func (oh *OrderHandlers) GetUserOrdersHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := helpers.GetUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		orders, err := oh.accrualService.GetUserOrders(ctx, userID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if len(orders) == 0 {
			ctx.Status(http.StatusNoContent)
			return
		}

		ctx.JSON(http.StatusOK, orders)
	}
}
