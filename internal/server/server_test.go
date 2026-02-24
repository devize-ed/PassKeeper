package server

import (
	"context"
	"net"
	"testing"

	"passKeper/internal/auth"
	"passKeper/internal/service/mocks"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewServer(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	storage := mocks.NewMockStorage(t)

	srv := NewServer(jwtManager, storage)
	require.NotNil(t, srv)
	require.NotNil(t, srv.AuthServer)
	require.NotNil(t, srv.ItemServer)
}

func TestServer_Serve(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	storage := mocks.NewMockStorage(t)

	tests := []struct {
		name      string
		addr      string
		cancelCtx bool
		wantErr   bool
		errStr    string
	}{
		{
			name:      "context cancellation",
			addr:      ":0",
			cancelCtx: true,
			wantErr:   false,
		},
		{
			name:    "listen error",
			addr:    "invalid-address",
			wantErr: true,
			errStr:  "listen",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewServer(jwtManager, storage)
			ctx, cancel := context.WithCancel(context.Background())
			if tc.cancelCtx {
				cancel()
			}
			defer cancel()

			err := srv.Serve(ctx, tc.addr, jwtManager)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errStr != "" {
					require.Contains(t, err.Error(), tc.errStr)
				}
				return
			}
			require.NoError(t, err)
		})
	}

	t.Run("accept connections", func(t *testing.T) {
		lis, err := net.Listen("tcp", ":0")
		require.NoError(t, err)
		addr := lis.Addr().String()
		lis.Close()

		srv := NewServer(jwtManager, storage)
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- srv.Serve(ctx, addr, jwtManager) }()

		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		conn.Close()

		cancel()
		require.NoError(t, <-done)
	})
}
