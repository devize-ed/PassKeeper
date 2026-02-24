package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadClientConfig(t *testing.T) {
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	defer func() { os.Args = originalArgs }()

	envKeys := []string{"SERVER_ADDRESS", "TOKEN_STORE_PATH", "LOG_LEVEL"}

	tests := []struct {
		name     string
		envVars  map[string]string
		expected ClientConfig
	}{
		{
			name:    "defaults when no env",
			envVars: map[string]string{},
			expected: ClientConfig{
				Address:        DefaultServerAddress,
				TokenStorePath: DefaultTokenStorePath,
				LogLevel:       DefaultLogLevel,
			},
		},
		{
			name: "env overrides",
			envVars: map[string]string{
				"SERVER_ADDRESS":   "custom:9090",
				"TOKEN_STORE_PATH": "/tmp/passkeeper-token",
			},
			expected: ClientConfig{
				Address:        "custom:9090",
				TokenStorePath: "/tmp/passkeeper-token",
				LogLevel:       DefaultLogLevel,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetViper()
			for _, k := range envKeys {
				_ = os.Unsetenv(k)
			}
			for k, v := range tc.envVars {
				assert.NoError(t, os.Setenv(k, v))
			}
			os.Args = []string{"cmd"}

			cfg, err := LoadClientConfig()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.Address, cfg.Address)
			assert.Equal(t, tc.expected.TokenStorePath, cfg.TokenStorePath)
			assert.Equal(t, tc.expected.LogLevel, cfg.LogLevel)
		})
	}
}
