// Package server provides the gRPC server implementation for PassKeeper.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"passKeper/internal/auth"
	"passKeper/internal/logger"
	"passKeper/internal/repository/db"
	"passKeper/internal/service"
	pb "passKeper/pkg/api"
	"time"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"

	"google.golang.org/grpc"
)

// AuthService interface provides the methods to interact with the auth service layer.
type AuthService interface {
	CreateUser(ctx context.Context, username string, password string) error
	LoginUser(ctx context.Context, username string, password string) (string, error)
}

// ItemService interface provides the methods to interact with the item service layer.
type ItemService interface {
	CreateItem(ctx context.Context, itemType int32, itemData []byte) (db.Item, error)
	UpdateItem(ctx context.Context, itemID string, itemType int32, itemData []byte, timestamp time.Time) (db.Item, error)
	DeleteItem(ctx context.Context, itemID string) error
	GetItem(ctx context.Context, id string) (db.Item, error)
	GetAllItems(ctx context.Context, itemType int32) ([]db.Item, error)
}

// NewServer creates a new server.
func NewServer(jwtManager *auth.JWTManager, storage service.Storage) *Server {
	// Create the auth and item services.
	authService := service.NewAuthService(storage, jwtManager)
	itemService := service.NewItemService(storage)
	// Create the server.
	return &Server{AuthServer: &AuthServer{AuthService: authService}, ItemServer: &ItemServer{ItemService: itemService}}
}

// Server holds the gRPC server and its auth/item service implementations.
type Server struct {
	AuthServer pb.PasskeeperAuthServiceServer
	ItemServer pb.PasskeeperItemServiceServer
	grpcServer *grpc.Server
}

// AuthServer implements the AuthService interface.
type AuthServer struct {
	pb.PasskeeperAuthServiceServer
	auth AuthService
}

// ItemServer implements the ItemService interface.
type ItemServer struct {
	pb.PasskeeperItemServiceServer
	ItemService ItemService
}

// Start starts the server.
func (s *Server) serverStart(host string) (net.Listener, error) {
	lis, err := net.Listen("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	// Create a new GRPC server.
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(grpcauth.UnaryServerInterceptor(authFunc(s.AuthServer.))),
	)
	// Register the services.
	pb.RegisterPasskeeperAuthServiceServer(s.grpcServer, s.AuthServer)
	pb.RegisterPasskeeperItemServiceServer(s.grpcServer, s.ItemServer)
	return lis, nil
}

// Stop stops the server.
func (s *Server) serverStop() {
	// Gracefully stop the server.
	logger.Log.Infof("Stopping the server")
	s.grpcServer.GracefulStop()
}

// Serve starts the GRPC server, blocks until ctx is cancelled, provides shutdown.
func (s *Server) Serve(ctx context.Context, host string, jwtm *auth.JWTManager) error {
	// Create a channel to receive the error from the server.
	errCh := make(chan error, 1)
	// Start the GRPC server
	lis, err := s.serverStart(host, jwtm)
	if err != nil {
		return err
	}
	go func() { errCh <- s.grpcServer.Serve(lis) }()
	// Wait for the context to be done or the server to be stopped.
	select {
	case <-ctx.Done():
		s.serverStop()
		return nil
	case err := <-errCh:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return fmt.Errorf("grpc serve: %w", err)
	}
}
