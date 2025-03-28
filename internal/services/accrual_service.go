package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
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
	httpClient       *resty.Client
	orderRepository  repository.OrderRepository
	pointsRepository repository.PointsRepository
	isRateLimited    bool
	jobsCh           chan workerJob
	buffer           *jobBuffer
	mutex            sync.Mutex
	logger           *zap.Logger
}

func NewAccrualService(config *config.Config, orderRepository repository.OrderRepository, pointsRepository repository.PointsRepository, logger *zap.Logger) AccrualService {
	client := resty.New()
	client.SetBaseURL(config.AccrualAddress)

	return &AccrualServiceImpl{
		httpClient:       client,
		orderRepository:  orderRepository,
		pointsRepository: pointsRepository,
		buffer:           NewJobBuffer(),
		logger:           logger,
		jobsCh:           make(chan workerJob),
	}
}

func (a *AccrualServiceImpl) StartWorker() {
	a.logger.Info("started worker")
	go func() {
		a.logger.Info("started goroutine")

		for job := range a.jobsCh {
			a.logger.Info("got job", zap.String("order_num", job.orderNum))
			if a.isRateLimited {
				a.logger.Info("put job in buffer")
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
			a.logger.Info("sent request", zap.String("url", resp.Request.URL))

			a.logger.Info("response from service", zap.Int("code", resp.StatusCode()), zap.String("body", string(resp.Body())))

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
				a.logger.Info("rate limited", zap.Duration("duration", rateLimitDuration), zap.String("header", resp.Header()["Retry-After"][0]))

				a.buffer.Add(job.orderNum, job)
				go a.waitForRateLimit(rateLimitDuration)

				continue
			}

			if resp.StatusCode() == http.StatusNoContent {
				result := workerResult{
					response: nil,
					err:      fmt.Errorf("204 no content"),
				}

				job.resultCh <- result
				continue
			}

			result := workerResult{
				response: &responseBody,
				err:      nil,
			}

			job.resultCh <- result
		}
		a.logger.Info("no jobs. exiting")
	}()
}

func (a *AccrualServiceImpl) waitForRateLimit(retryAfter time.Duration) {
	a.logger.Info("start sleeping")

	time.Sleep(retryAfter)

	a.mutex.Lock()
	a.isRateLimited = false
	a.mutex.Unlock()
	a.logger.Info("unlock rate limit")

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

	for i, order := range orders {
		if !isOrderAccrualCalculated(order.AccrualStatus) {
			err := a.QueueStatusUpdate(ctx, order.OrderNum) // queue update
			if err != nil {
				return nil, err
			}
		}

		if order.AccrualStatus == models.AccrualStatusRegistered {
			orders[i].AccrualStatus = models.AccrualStatusNew
		}
	}

	return orders, nil
}

func (a *AccrualServiceImpl) QueueStatusUpdate(ctx context.Context, orderNum string) error {
	order, err := a.orderRepository.GetOrder(ctx, orderNum)

	if err != nil {
		return err
	}
	a.logger.Info("order status", zap.String("status", string(order.AccrualStatus)))
	a.logger.Info("is calculated", zap.Bool("result", isOrderAccrualCalculated(order.AccrualStatus)))

	if !isOrderAccrualCalculated(order.AccrualStatus) {
		job := workerJob{
			orderNum: orderNum,
			resultCh: make(chan workerResult),
		}

		go func() {
			a.jobsCh <- job
			a.logger.Info("put job", zap.String("order_num", job.orderNum))
			result := <-job.resultCh
			//a.logger.Info("got job result", zap.String("accrual", result.response.Accrual.String()), zap.Error(result.err))
			if result.err != nil {
				a.logger.Info("error", zap.Error(result.err))
				return
			}
			a.logger.Info("statuses", zap.String("old", string(order.AccrualStatus)), zap.String("new", string(result.response.Status)))

			if order.AccrualStatus != result.response.Status {
				a.orderRepository.UpdateOrderStatus(context.Background(), orderNum, result.response.Status, result.response.Accrual) // ctx? err?
				a.logger.Info("is calculated", zap.Bool("result", isOrderAccrualCalculated(result.response.Status)))
				if isOrderAccrualCalculated(result.response.Status) {
					a.pointsRepository.AddPoints(context.Background(), order.UserID, *result.response.Accrual)
				}
			}
		}()
	}

	return nil
}

func isOrderAccrualCalculated(accrualStatus models.AccrualStatus) bool {
	return accrualStatus == models.AccrualStatusInvalid || accrualStatus == models.AccrualStatusProcessed
}
