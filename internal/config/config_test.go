package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const programName = "test"

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantConfig Config
	}{
		{
			"empty args",
			[]string{programName},
			*NewConfig(),
		},
		{
			"only RunAddress",
			[]string{programName, "-a", ":8888"},
			*NewConfig(WithRunAddress(":8888")),
		},
		{
			"only LogLevel",
			[]string{programName, "-l", "info"},
			*NewConfig(WithLogLevel("info")),
		},
		{
			"only accrual address",
			[]string{programName, "-r", ":8080"},
			*NewConfig(WithAccrualAddress(":8080")),
		},
		{
			"only database URI",
			[]string{programName, "-d", "postgresql://user@localhost/db"},
			*NewConfig(WithDatabaseUri("postgresql://user@localhost/db")),
		},
		{
			"only token secret",
			[]string{programName, "-t", "supersecretkey"},
			*NewConfig(WithTokenSecret("supersecretkey")),
		},
		{
			"full args",
			[]string{programName, "-a", ":8888", "-d", "postgresql://user@localhost/db", "-l", "debug", "-r", ":8080", "-t", "supersecretkey"},
			*NewConfig(WithRunAddress(":8888"), WithLogLevel("debug"), WithAccrualAddress(":8080"), WithDatabaseUri("postgresql://user@localhost/db"), WithTokenSecret("supersecretkey")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualConfig, err := ParseArgs(tt.args[0], tt.args[1:])

			require.NoError(t, err)
			assert.Equal(t, &tt.wantConfig, actualConfig)
		})
	}
}

func TestParseArgsErr(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr error
	}{
		{
			"invalid AppRunAddress",
			[]string{programName, "-a", "test:onetwothree"},
			ErrInvalidRunAddress,
		},
		{
			"invalid AccrualAddress",
			[]string{programName, "-r", "test:onetwothree"},
			ErrInvalidAccrualAddress,
		},
		{
			"invalid Database URI",
			[]string{programName, "-d", ""},
			ErrInvalidDatabaseUri,
		},
		{
			"invalid LogLevel",
			[]string{programName, "-l", "debug123"},
			ErrInvalidLogLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args[0], tt.args[1:])
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
