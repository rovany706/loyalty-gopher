package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	Registered OrderStatus = "REGISTERED"
	Processing OrderStatus = "PROCESSING"
	Invalid    OrderStatus = "INVALID"
	Processed  OrderStatus = "PROCESSED"
)

var (
	ErrRateLimit = errors.New("too many requests to accrual service")
)

type AccrualServiceResponse struct {
	Order   string               `json:"order"`
	Status  models.AccrualStatus `json:"status"`
	Accrual *decimal.Decimal     `json:"accrual,omitempty"`
}

type AccrualService interface {
	GetOrderStatus(ctx context.Context, orderNum string) (*AccrualServiceResponse, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
}

type AccrualServiceImpl struct {
	httpClient      *resty.Client
	orderRepository repository.OrderRepository
}

func NewAccrualService(config *config.Config, orderRepository repository.OrderRepository) AccrualService {
	client := resty.New()
	client.SetBaseURL(config.AccrualAddress)

	return &AccrualServiceImpl{
		httpClient:      client,
		orderRepository: orderRepository,
	}
}

func (a *AccrualServiceImpl) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	orders, err := a.orderRepository.GetUserOrders(ctx, userID)

	if err != nil {
		return nil, err
	}

	for i, order := range orders {
		if order.AccrualStatus != models.AccrualStatusInvalid && order.AccrualStatus != models.AccrualStatusProcessed {
			resp, err := a.GetOrderStatus(ctx, order.OrderNum)
			if err != nil {
				if errors.Is(err, ErrRateLimit) {
					continue
				}

				return nil, err
			}

			if resp.Status != order.AccrualStatus {
				orders[i].AccrualStatus = resp.Status
				orders[i].Accrual = resp.Accrual
				a.orderRepository.UpdateOrderStatus(ctx, orders[i].OrderNum, resp.Status, resp.Accrual)
			}
		}
	}

	return orders, nil
}

func (a *AccrualServiceImpl) GetOrderStatus(ctx context.Context, orderNum string) (*AccrualServiceResponse, error) {
	responseBody := &AccrualServiceResponse{}
	resp, err := a.httpClient.R().
		SetResult(responseBody).
		SetPathParam("orderNum", orderNum).
		Get("/api/orders/{orderNum}")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		return nil, ErrRateLimit
	}

	return responseBody, nil
}
