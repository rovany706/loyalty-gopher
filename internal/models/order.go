package models

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type OrderTime time.Time

func (t OrderTime) MarshalJSON() ([]byte, error) {
	formattedTime := fmt.Sprintf(`"%s"`, time.Time(t).Format(time.RFC3339))

	return []byte(formattedTime), nil
}

type AccrualStatus string

const (
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
)

type Order struct {
	UserID        int              `json:"-"`
	OrderNum      string           `json:"number"`
	UploadedAt    OrderTime        `json:"uploaded_at"`
	AccrualStatus AccrualStatus    `json:"status"`
	Accrual       *decimal.Decimal `json:"accrual,omitempty"`
}

type GetUserOrdersResponse []Order
