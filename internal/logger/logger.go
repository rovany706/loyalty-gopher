package logger

import (
	"go.uber.org/zap"
)

func NewLogger(logLevel string) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(logLevel)

	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	zapLogger, err := cfg.Build()

	if err != nil {
		return nil, err
	}

	return zapLogger, nil
}
