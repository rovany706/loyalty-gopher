package server

import (
	"errors"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/database"
	"github.com/rovany706/loyalty-gopher/internal/handlers"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"github.com/rovany706/loyalty-gopher/internal/routes"
	"github.com/rovany706/loyalty-gopher/internal/services"
	"go.uber.org/zap"
)

type Server struct {
	config           *config.Config
	logger           *zap.Logger
	database         *database.Database
	userRepository   repository.UserRepository
	orderRepository  repository.OrderRepository
	pointsRepository repository.PointsRepository
	tokenManager     auth.TokenManager
	accrualService   services.AccrualService
}

func NewServer(config *config.Config, logger *zap.Logger, database *database.Database) (*Server, error) {
	userRepository := repository.NewDBUserRepository(database)
	orderRepository := repository.NewDBOrderRepository(database, logger)
	pointsRepository := repository.NewDBPointsRepository(database, logger)
	tokenManager, err := auth.NewJWTTokenManager([]byte(config.TokenSecret))
	accrualService := services.NewAccrualService(config, orderRepository, pointsRepository, logger)

	if err != nil {
		return nil, err
	}

	return &Server{
		config:           config,
		logger:           logger,
		database:         database,
		tokenManager:     tokenManager,
		userRepository:   userRepository,
		orderRepository:  orderRepository,
		pointsRepository: pointsRepository,
		accrualService:   accrualService,
	}, nil
}

func (s *Server) Run() (err error) {
	s.accrualService.StartWorker()
	defer s.accrualService.StopWorker()
	defer func() {
		err = errors.Join(err, s.database.Close())
	}()

	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	routes.RegisterAuthHandlers(r, handlers.NewAuthHandlers(s.userRepository, s.tokenManager))
	routes.RegisterOrderHandlers(r, handlers.NewOrderHandlers(s.orderRepository, s.accrualService), s.tokenManager)
	routes.RegisterPointsHandlers(r, handlers.NewPointsHandlers(s.pointsRepository), s.tokenManager)

	return r.Run(s.config.RunAddress)
}
