package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthServer_AuthFuncOverride(t *testing.T) {
	ctx := context.Background()
	srv := &AuthServer{}

	outCtx, err := srv.AuthFuncOverride(ctx, "/passkeeper.PasskeeperAuthService/Login")
	assert.NoError(t, err)
	assert.NotNil(t, outCtx)

	outCtx, err = srv.AuthFuncOverride(ctx, "/passkeeper.PasskeeperAuthService/Register")
	assert.NoError(t, err)
	assert.NotNil(t, outCtx)
}
