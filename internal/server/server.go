package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rovany706/loyalty-gopher/internal/auth"
	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/database"
	"github.com/rovany706/loyalty-gopher/internal/handlers"
	"github.com/rovany706/loyalty-gopher/internal/repository"
	"go.uber.org/zap"
)

type Server struct {
	config         *config.Config
	logger         *zap.Logger
	database       *database.Database
	userRepository repository.UserRepository
	tokenManager   auth.TokenManager
}

func NewServer(config *config.Config, logger *zap.Logger, database *database.Database) (*Server, error) {
	userRepository := repository.NewDBUserRepository(database)
	tokenManager, err := auth.NewJWTTokenManager([]byte(config.TokenSecret))
	if err != nil {
		return nil, err
	}

	return &Server{
		config:         config,
		logger:         logger,
		database:       database,
		tokenManager:   tokenManager,
		userRepository: userRepository,
	}, nil
}

func (s *Server) Run() error {
	return s.getRouter().Run(s.config.RunAddress)
}

func (s *Server) getRouter() *gin.Engine {
	r := gin.Default()

	userAPIGroup := r.Group("/api/user")
	{
		userAPIGroup.POST("/register", handlers.RegisterHandler(s.userRepository, s.tokenManager))
		userAPIGroup.POST("/login", handlers.LoginHandler(s.userRepository, s.tokenManager))
	}

	return r
}
