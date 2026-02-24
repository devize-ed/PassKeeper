package grpcclient

import (
	"context"
	"fmt"
	pb "passKeper/pkg/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client is the gRPC client for the passkeeper service.
type Client struct {
	conn *grpc.ClientConn
	pb.PasskeeperAuthServiceClient
	pb.PasskeeperItemServiceClient
}

// NewClient creates a new gRPC client. Pass nil for store to skip auth token injection.
func NewClient(host string, store TokenStore) (*Client, error) {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(authInterceptor(store)))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}
	return &Client{
		conn:                        conn,
		PasskeeperAuthServiceClient: pb.NewPasskeeperAuthServiceClient(conn),
		PasskeeperItemServiceClient: pb.NewPasskeeperItemServiceClient(conn),
	}, nil
}

// Close closes the gRPC client.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Register registers a new user.
func (c *Client) Register(ctx context.Context, username, password string) error {
	_, err := c.PasskeeperAuthServiceClient.Register(ctx, &pb.RegisterRequest{Username: username, Password: password})
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	return nil
}

// Login logs in a user.
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	response, err := c.PasskeeperAuthServiceClient.Login(ctx, &pb.LoginRequest{Username: username, Password: password})
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	return response.Token, nil
}

// CreateItem creates a new item.
func (c *Client) CreateItem(ctx context.Context, itemType pb.ItemType, itemData *pb.ItemData) error {
	_, err := c.PasskeeperItemServiceClient.CreateItem(ctx, &pb.CreateItemRequest{Type: itemType, Data: itemData})
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// UpdateItem updates an item.
func (c *Client) UpdateItem(ctx context.Context, itemID string, itemType pb.ItemType, itemData *pb.ItemData, updatedAt *timestamppb.Timestamp) error {
	_, err := c.PasskeeperItemServiceClient.UpdateItem(ctx, &pb.UpdateItemRequest{Id: itemID, Type: itemType, Data: itemData, UpdatedAt: updatedAt})
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

// DeleteItem deletes an item.
func (c *Client) DeleteItem(ctx context.Context, itemID string) error {
	_, err := c.PasskeeperItemServiceClient.DeleteItem(ctx, &pb.DeleteItemRequest{Id: itemID})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// GetItem gets an item.
func (c *Client) GetItem(ctx context.Context, itemID string) (*pb.Item, error) {
	response, err := c.PasskeeperItemServiceClient.GetItem(ctx, &pb.GetItemRequest{Id: itemID})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	return response.Item, nil
}

// ListItems lists all items.
func (c *Client) ListItems(ctx context.Context, itemType pb.ItemType) ([]*pb.Item, error) {
	response, err := c.PasskeeperItemServiceClient.ListItems(ctx, &pb.ListItemsRequest{Type: itemType})
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	return response.GetItems(), nil
}
