package models

import (
	"github.com/shopspring/decimal"
)

type GetUserBalanceResponse struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type WithdrawHistoryEntry struct {
	OrderNum    string          `json:"order"`
	WithdrawSum decimal.Decimal `json:"sum"`
	ProcessedAt RFC3339Time     `json:"processed_at"`
}

type WithdrawUserPointsRequest struct {
	OrderNum    string          `json:"order"`
	WithdrawSum decimal.Decimal `json:"sum"`
}
