package repository

import (
	"context"
	"errors"

	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/shopspring/decimal"
)

var (
	ErrOrderConflict = errors.New("order already exists")
)

type OrderRepository interface {
	GetOrder(ctx context.Context, orderNum string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
	AddOrder(ctx context.Context, userID int, orderNum string) error
	UpdateOrderStatus(ctx context.Context, orderNum string, newAccrualStatus models.AccrualStatus, accrualAmount *decimal.Decimal) error
}
