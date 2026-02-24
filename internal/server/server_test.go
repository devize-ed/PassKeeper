package server

import (
	"testing"

	"passKeper/internal/auth"
	"passKeper/internal/service/mocks"

	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	storage := mocks.NewMockStorage(t)

	srv := NewServer(jwtManager, storage)
	require.NotNil(t, srv)
	require.NotNil(t, srv.AuthServer)
	require.NotNil(t, srv.ItemServer)
}
