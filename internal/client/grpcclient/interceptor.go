package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TokenStore interface provides the methods to interact with the token store.
type TokenStore interface {
	Load(ctx context.Context) (string, error)
}

// authInterceptor returns a unary client interceptor that injects the auth token from the store.
func authInterceptor(store TokenStore) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string,
		req interface{}, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if store != nil {
			token, _ := store.Load(ctx)
			if token != "" {
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
