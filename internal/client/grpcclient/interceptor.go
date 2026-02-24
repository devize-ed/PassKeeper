package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// LoginMethod and RegisterMethod are the grpcmethods that do not require authentication.
const (
	LoginMethod    = "/passkeeper.PasskeeperAuthService/Login"
	RegisterMethod = "/passkeeper.PasskeeperAuthService/Register"
)

// TokenStore interface provides the methods to interact with the token store.
type TokenStore interface {
	Load() (string, error)
}

// authInterceptor returns a unary client interceptor that injects the auth token from the store.
func authInterceptor(ts TokenStore) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string,
		req interface{}, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if ts != nil {
			// skip authentication for login and register methods
			if method == LoginMethod || method == RegisterMethod {
				return invoker(ctx, method, req, reply, cc, opts...)
			}
			// load the token from the store
			token, err := ts.Load()
			if err != nil {
				return fmt.Errorf("authentication error: %w", err)
			}
			// add the token to the context
			if token != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
