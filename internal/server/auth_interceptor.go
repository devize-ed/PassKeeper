package server

import (
	"context"

	"passKeper/internal/auth"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authFunc is a function that authenticates the user and returns the context.
func authFunc(jwtm *auth.JWTManager) grpcauth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		// Get the token from the context
		token, err := grpcauth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		// Parse the token and get the user ID
		userID, err := jwtm.ParseToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}
		// Add the user ID to the context
		ctx = auth.WithUserID(ctx, userID)
		return ctx, nil
	}
}

// AuthFuncOverride is a function that overrides the authentication function to pass the auth for login and register services.
func (s *AuthServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
