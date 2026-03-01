package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveOriginalState saves the original state to restore after tests.
func saveOriginalState() ([]string, *pflag.FlagSet) {
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	originalFlags := pflag.CommandLine
	return originalArgs, originalFlags
}

// restoreOriginalState restores the original state.
func restoreOriginalState(originalArgs []string, originalFlags *pflag.FlagSet) {
	os.Args = originalArgs
	pflag.CommandLine = originalFlags
}

func resetViper() {
	viper.Reset()
}

func TestLoadServerConfig(t *testing.T) {
	originalArgs, originalFlags := saveOriginalState()
	defer restoreOriginalState(originalArgs, originalFlags)

	envKeys := []string{"SERVER_ADDRESS", "DATABASE_DSN", "AUTH_SECRET", "LOG_LEVEL"}

	tests := []struct {
		name     string
		envVars  map[string]string
		args     []string
		expected *ServerConfig
		wantErr  bool
	}{
		{
			name: "defaults",
			envVars: map[string]string{
				"DATABASE_DSN": "test_dsn",
				"AUTH_SECRET":  "test_secret",
			},
			args: []string{},
			expected: &ServerConfig{
				Address:  DefaultHost,
				DSN:      "test_dsn",
				AuthKey:  "test_secret",
				LogLevel: DefaultLogLevel,
			},
			wantErr: false,
		},
		{
			name: "env overrides",
			envVars: map[string]string{
				"SERVER_ADDRESS": "envhost:7777",
				"DATABASE_DSN":   "test_dsn",
				"AUTH_SECRET":    "test_secret",
				"LOG_LEVEL":      "error",
			},
			args: []string{},
			expected: &ServerConfig{
				Address:  "envhost:7777",
				DSN:      "test_dsn",
				AuthKey:  "test_secret",
				LogLevel: "error",
			},
			wantErr: false,
		},
		{
			name: "flags override env",
			envVars: map[string]string{
				"SERVER_ADDRESS": "envhost:9000",
				"DATABASE_DSN":   "any_dsn",
				"AUTH_SECRET":    "any_secret",
				"LOG_LEVEL":      "warn",
			},
			args: []string{"-a=:9999", "-d=test_dsn", "-k=test_secret", "-l=info"},
			expected: &ServerConfig{
				Address:  ":9999",
				DSN:      "test_dsn",
				AuthKey:  "test_secret",
				LogLevel: "info",
			},
			wantErr: false,
		},
		{
			name: "missing auth_key",
			envVars: map[string]string{
				"DATABASE_DSN": "test_dsn",
			},
			args:     []string{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "missing dsn",
			envVars: map[string]string{
				"AUTH_SECRET": "test_secret",
			},
			args:     []string{},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetViper()
			pflag.CommandLine = pflag.NewFlagSet(tc.name, pflag.ContinueOnError)

			for _, k := range envKeys {
				_ = os.Unsetenv(k)
			}
			for k, v := range tc.envVars {
				assert.NoError(t, os.Setenv(k, v))
			}
			os.Args = append([]string{"cmd"}, tc.args...)

			cfg, err := LoadServerConfig()
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, cfg)
				assert.Equal(t, tc.expected.Address, cfg.Address)
				assert.Equal(t, tc.expected.DSN, cfg.DSN)
				assert.Equal(t, tc.expected.AuthKey, cfg.AuthKey)
				assert.Equal(t, tc.expected.LogLevel, cfg.LogLevel)
			}
		})
	}
}

func Test_validateServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ServerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &ServerConfig{
				AuthKey: "test_secret",
				DSN:     "test_dsn",
			},
			wantErr: false,
		},
		{
			name: "missing auth key",
			cfg: &ServerConfig{
				DSN: "test_dsn",
			},
			wantErr: true,
		},
		{
			name: "missing dsn",
			cfg: &ServerConfig{
				AuthKey: "test_secret",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateServerConfig(tt.cfg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateServerConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateServerConfig() succeeded unexpectedly")
			}
		})
	}
}
