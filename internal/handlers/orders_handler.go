package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/middleware"
	"github.com/rovany706/loyalty-gopher/internal/repository"
)

type OrderHandlers struct {
	orderRepository repository.OrderRepository
}

func NewOrderHandlers(orderRepository repository.OrderRepository) *OrderHandlers {
	return &OrderHandlers{
		orderRepository: orderRepository,
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

		ok, err := luhnCheck(orderNum)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !ok {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}

		userID, ok := getUserIDFromContext(ctx)
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

		ctx.Status(http.StatusCreated)
	}
}

func (oh *OrderHandlers) GetUserOrdersHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := getUserIDFromContext(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		orders, err := oh.orderRepository.GetUserOrders(ctx, userID)
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

func getUserIDFromContext(ctx *gin.Context) (int, bool) {
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

func convertStringToIntSlice(str string) ([]int, error) {
	nums := make([]int, len(str))
	for i, s := range str {
		nums[i] = int(s - '0')
	}

	return nums, nil
}

func luhnCheck(orderNum string) (bool, error) {
	orderNums, err := convertStringToIntSlice(orderNum)
	if err != nil {
		return false, err
	}

	if len(orderNums) == 0 {
		return false, nil
	}

	sum := 0
	parity := len(orderNums) % 2

	for i := range orderNums {
		digit := orderNums[i]
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0, nil
}
