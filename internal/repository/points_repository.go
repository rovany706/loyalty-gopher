package repository

import (
	"context"
	"errors"

	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/shopspring/decimal"
)

var ErrNotEnoughPoints = errors.New("not enough points to withdraw")

type PointsRepository interface {
	GetUserBalance(ctx context.Context, userID int) (decimal.Decimal, error)
	GetUserWithdrawalHistory(ctx context.Context, userID int) ([]models.WithdrawHistoryEntry, error)
	WithdrawPoints(ctx context.Context, userID int, orderNum string, amount decimal.Decimal) error
	AddPoints(ctx context.Context, userID int, amount decimal.Decimal) error
}
