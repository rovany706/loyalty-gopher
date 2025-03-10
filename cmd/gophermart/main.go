package main

import (
	"context"
	"os"
	"time"

	"github.com/rovany706/loyalty-gopher/internal/config"
	"github.com/rovany706/loyalty-gopher/internal/database"
	"github.com/rovany706/loyalty-gopher/internal/logger"
	"github.com/rovany706/loyalty-gopher/internal/server"
	"go.uber.org/zap"
)

func main() {
	config, err := config.ParseArgs(os.Args[0], os.Args[1:])

	if err != nil {
		panic(err)
	}

	logger, err := logger.NewLogger(config.LogLevel)

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	database, err := database.InitConnection(ctx, config.DatabaseUri)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	err = database.RunMigrations(ctx)
	if err != nil {
		logger.Fatal("error running migrations", zap.Error(err))
	}

	server, err := server.NewServer(config, logger, database)
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	if err := server.Run(); err != nil {
		logger.Fatal("error when running server", zap.Error(err))
	}
}
