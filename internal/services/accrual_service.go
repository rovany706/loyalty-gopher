package services

type OrderStatus int

const (
	New OrderStatus = iota
	Processing
	Invalid
	Processed
)

type AccrualService interface {
	GetOrderStatus(orderNum string) error
}

type AccrualServiceImpl struct {
}
