package services

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/helpers"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var (
	errRateLimit = errors.New("rate limit")
	errNoContent = errors.New("no content")
)

type accrualServiceResponse struct {
	OrderNum string               `json:"order"`
	Status   models.AccrualStatus `json:"status"`
	Accrual  *decimal.Decimal     `json:"accrual,omitempty"`
}

type AccrualService interface {
	StartWorker()
	StopWorker()
	QueueStatusUpdate(order models.Order)
}

type workerJob struct {
	orderNum string
	resultCh chan workerResult
}

type workerResult struct {
	response *accrualServiceResponse
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
	go func() {
		a.logger.Info("started worker")

		for job := range a.jobsCh {
			a.logger.Info("got job", zap.String("order_num", job.orderNum))

			if a.isRateLimited {
				a.logger.Info("put job in buffer")
				a.buffer.Add(job.orderNum, job)

				result := workerResult{
					response: nil,
					err:      errRateLimit,
				}
				job.resultCh <- result
				continue
			}

			responseBody := accrualServiceResponse{}
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

			if err := a.checkResponse(resp, job); err != nil {
				result := workerResult{
					response: nil,
					err:      err,
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
	}()
}

func (a *AccrualServiceImpl) checkResponse(resp *resty.Response, job workerJob) error {
	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusTooManyRequests:
		err := a.handleTooManyRequests(resp, job)

		if err != nil {
			return err
		}

		return errRateLimit

	case http.StatusNoContent:
		return errNoContent
	}

	return nil
}

func (a *AccrualServiceImpl) handleTooManyRequests(resp *resty.Response, job workerJob) error {
	rateLimitDuration, err := time.ParseDuration(resp.Header()["Retry-After"][0] + "s")
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.isRateLimited = true
	a.mutex.Unlock()
	a.logger.Info("rate limited", zap.Duration("duration", rateLimitDuration), zap.String("header", resp.Header()["Retry-After"][0]))

	a.buffer.Add(job.orderNum, job)
	go a.waitForRateLimit(rateLimitDuration)

	return nil
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

func (a *AccrualServiceImpl) QueueStatusUpdate(order models.Order) {
	if helpers.IsOrderAccrualCalculated(order.AccrualStatus) {
		return
	}

	job := workerJob{
		orderNum: order.OrderNum,
		resultCh: make(chan workerResult),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	go func() {
		a.logger.Info("put new job", zap.String("order_num", job.orderNum))
		a.jobsCh <- job

		result := <-job.resultCh
		if result.err != nil {
			a.logger.Info("error processing job", zap.Error(result.err))
			return
		}

		err := a.orderRepository.UpdateOrderStatus(ctx, order.OrderNum, result.response.Status, result.response.Accrual)
		if err != nil {
			a.logger.Info("error updating order status", zap.Error(err))
		}
	}()
}

func (a *AccrualServiceImpl) StopWorker() {
	close(a.jobsCh)
}
