package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{
			name:     "valid log level",
			logLevel: "debug",
			wantErr:  false,
		},
		{
			name:     "invalid log level",
			logLevel: "debug level, please",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.logLevel)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, logger)
		})
	}
}
