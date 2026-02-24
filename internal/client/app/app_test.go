package app

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	client "passKeper/internal/client/grpcclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAppTokenStore implements TokenStore for tests.
type mockAppTokenStore struct {
	token string
	err   error
}

func (m *mockAppTokenStore) Save(token string) error {
	return nil
}

func (m *mockAppTokenStore) Load() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestNewApp(t *testing.T) {
	tests := []struct {
		name       string
		grpcClient *client.Client
		ts         TokenStore
		wantErr    bool
	}{
		{
			name:       "with client and store",
			grpcClient: mustNewClient(t),
			ts:         &mockAppTokenStore{token: "t"},
			wantErr:    false,
		},
		{
			name:       "nil client nil store",
			grpcClient: nil,
			ts:         nil,
			wantErr:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.grpcClient != nil {
				defer tc.grpcClient.Close()
			}
			got, err := NewApp(tc.grpcClient, tc.ts)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func mustNewClient(t *testing.T) *client.Client {
	t.Helper()
	c, err := client.NewClient("localhost:50051", nil)
	require.NoError(t, err)
	return c
}

func TestApp_Close(t *testing.T) {
	c := mustNewClient(t)
	app, err := NewApp(c, nil)
	require.NoError(t, err)
	err = app.Close()
	require.NoError(t, err)
}

func TestApp_requireAuthentication(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		loadErr     error
		wantErr     bool
		errContains string
	}{
		{
			name:    "has token",
			token:   "valid-token",
			loadErr: nil,
			wantErr: false,
		},
		{
			name:        "empty token",
			token:       "",
			loadErr:     nil,
			wantErr:     true,
			errContains: "authentication required",
		},
		{
			name:        "store load error",
			token:       "",
			loadErr:     errors.New("load failed"),
			wantErr:     true,
			errContains: "authentication required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{
				ts:  &mockAppTokenStore{token: tc.token, err: tc.loadErr},
				in:  bufio.NewReader(bytes.NewReader(nil)),
				out: bufio.NewWriter(&bytes.Buffer{}),
			}
			err := a.requireAuthentication()
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
