package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	TokenSecret    string `env:"TOKEN_SECRET"`
}

const (
	defaultRunAddress     = ":8081"
	defaultDatabaseURI    = "postgresql://app:example@localhost:5432/gophermartdb"
	defaultAccrualAddress = "http://localhost:8080"
	defaultLogLevel       = "info"
	defaultTokenSecret    = "secret"
)

var (
	ErrInvalidRunAddress     = errors.New("invalid run address")
	ErrInvalidDatabaseURI    = errors.New("invalid database URI")
	ErrInvalidAccrualAddress = errors.New("invalid accrual address")
	ErrInvalidLogLevel       = errors.New("invalid log level")
)

type Option func(config *Config)

func WithRunAddress(runAddress string) Option {
	return func(config *Config) {
		config.RunAddress = runAddress
	}
}

func WithDatabaseURI(databaseURI string) Option {
	return func(config *Config) {
		config.DatabaseURI = databaseURI
	}
}

func WithAccrualAddress(accrualAddress string) Option {
	return func(config *Config) {
		config.AccrualAddress = accrualAddress
	}
}

func WithLogLevel(logLevel string) Option {
	return func(config *Config) {
		config.LogLevel = logLevel
	}
}

func WithTokenSecret(tokenSecret string) Option {
	return func(config *Config) {
		config.TokenSecret = tokenSecret
	}
}

func NewConfig(opts ...Option) *Config {
	config := &Config{
		RunAddress:     defaultRunAddress,
		DatabaseURI:    defaultDatabaseURI,
		AccrualAddress: defaultAccrualAddress,
		LogLevel:       defaultLogLevel,
		TokenSecret:    defaultTokenSecret,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func ParseArgs(programName string, args []string) (config *Config, err error) {
	config = new(Config)
	flags := flag.NewFlagSet(programName, flag.ExitOnError)

	flags.StringVar(&config.RunAddress, "a", defaultRunAddress, fmt.Sprintf("address and port to run server (default: %s)", defaultRunAddress))
	flags.StringVar(&config.DatabaseURI, "d", defaultDatabaseURI, fmt.Sprintf("database DSN (default: %s)", defaultDatabaseURI))
	flags.StringVar(&config.AccrualAddress, "r", defaultAccrualAddress, fmt.Sprintf("address and port of accrual system (default: %s)", defaultAccrualAddress))
	flags.StringVar(&config.LogLevel, "l", defaultLogLevel, fmt.Sprintf("log level (default: %s)", defaultLogLevel))
	flags.StringVar(&config.TokenSecret, "t", defaultTokenSecret, fmt.Sprintf("token secret (default: %s)", defaultTokenSecret))

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, err
	}

	log.Printf("Parsed app config: %+v\n", config)

	if err := validateParsedArgs(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateParsedArgs(config *Config) error {
	if _, err := net.ResolveTCPAddr("tcp", config.RunAddress); err != nil {
		return ErrInvalidRunAddress
	}

	if !isURL(config.AccrualAddress) {
		return ErrInvalidAccrualAddress
	}

	if _, err := zap.ParseAtomicLevel(config.LogLevel); err != nil {
		return ErrInvalidLogLevel
	}

	if config.DatabaseURI == "" {
		return ErrInvalidDatabaseURI
	}

	return nil
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
