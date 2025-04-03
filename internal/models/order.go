package models

import (
	"github.com/shopspring/decimal"
)

type AccrualStatus string

const (
	AccrualStatusNew        AccrualStatus = "NEW"
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
)

type Order struct {
	UserID        int              `json:"-"`
	OrderNum      string           `json:"number"`
	UploadedAt    RFC3339Time      `json:"uploaded_at"`
	AccrualStatus AccrualStatus    `json:"status"`
	Accrual       *decimal.Decimal `json:"accrual,omitempty"`
}

type GetUserOrdersResponse []Order
