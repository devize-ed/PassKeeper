package logger_test

import (
	"passKeper/internal/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "valid level",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := logger.Initialize(tc.level)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}
