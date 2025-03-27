package services

import (
	"context"
	"net/http"
	"sync"
	"time"

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

type AccrualServiceResponse struct {
	OrderNum string               `json:"order"`
	Status   models.AccrualStatus `json:"status"`
	Accrual  *decimal.Decimal     `json:"accrual,omitempty"`
}

type AccrualService interface {
	StartWorker()
	QueueStatusUpdate(ctx context.Context, orderNum string) error
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
}

type workerJob struct {
	orderNum string
	resultCh chan workerResult
}

type workerResult struct {
	response *AccrualServiceResponse
	err      error
}

type AccrualServiceImpl struct {
	httpClient      *resty.Client
	orderRepository repository.OrderRepository
	isRateLimited   bool
	jobsCh          chan workerJob
	buffer          *jobBuffer
	mutex           sync.Mutex
}

func NewAccrualService(config *config.Config, orderRepository repository.OrderRepository) AccrualService {
	client := resty.New()
	client.SetBaseURL(config.AccrualAddress)

	return &AccrualServiceImpl{
		httpClient:      client,
		orderRepository: orderRepository,
		buffer:          NewJobBuffer(),
	}
}

func (a *AccrualServiceImpl) StartWorker() {
	go func() {
		for job := range a.jobsCh {
			if a.isRateLimited {
				a.buffer.Add(job.orderNum, job)
				continue
			}

			responseBody := AccrualServiceResponse{}
			resp, err := a.httpClient.R().
				SetResult(&responseBody).
				SetPathParam("orderNum", job.orderNum).
				Get("/api/orders/{orderNum}")

			if err != nil {
				result := workerResult{
					response: nil,
					err:      err,
				}
				job.resultCh <- result
				continue
			}

			if resp.StatusCode() == http.StatusTooManyRequests {
				rateLimitDuration, err := time.ParseDuration(resp.Header()["Retry-After"][0] + "s")
				if err != nil {
					result := workerResult{
						response: nil,
						err:      err,
					}
					job.resultCh <- result
				}

				a.mutex.Lock()
				a.isRateLimited = true
				a.mutex.Unlock()

				a.buffer.Add(job.orderNum, job)
				go a.waitForRateLimit(rateLimitDuration)

				continue
			}

			result := workerResult{
				response: &responseBody,
				err:      nil,
			}

			job.resultCh <- result
		}
	}()
}

func (a *AccrualServiceImpl) waitForRateLimit(retryAfter time.Duration) {
	time.Sleep(retryAfter)

	a.mutex.Lock()
	a.isRateLimited = false
	a.mutex.Unlock()

	jobs := a.buffer.Flush()

	for _, job := range jobs {
		a.jobsCh <- job
	}
}

func (a *AccrualServiceImpl) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	orders, err := a.orderRepository.GetUserOrders(ctx, userID)

	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		if !isOrderAccrualCalculated(order) {
			err := a.QueueStatusUpdate(ctx, order.OrderNum) // queue update
			if err != nil {
				return nil, err
			}
		}
	}

	return orders, nil
}

func (a *AccrualServiceImpl) QueueStatusUpdate(ctx context.Context, orderNum string) error {
	order, err := a.orderRepository.GetOrder(ctx, orderNum)

	if err != nil {
		return err
	}

	if !isOrderAccrualCalculated(*order) {
		job := workerJob{
			orderNum: orderNum,
			resultCh: make(chan workerResult),
		}

		go func() {
			a.jobsCh <- job
			result := <-job.resultCh
			if order.AccrualStatus != result.response.Status {
				a.orderRepository.UpdateOrderStatus(context.Background(), orderNum, result.response.Status, result.response.Accrual) // ctx? err?
			}
		}()
	}

	return nil
}

func isOrderAccrualCalculated(order models.Order) bool {
	return order.AccrualStatus != models.AccrualStatusInvalid && order.AccrualStatus != models.AccrualStatusProcessed
}
