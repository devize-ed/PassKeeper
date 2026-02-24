package grpcclient_test

import (
	"context"
	"errors"
	"testing"

	"passKeper/internal/client/grpcclient"
	"passKeper/internal/client/grpcclient/mocks"
	pb "passKeper/pkg/api"
	pbMocks "passKeper/pkg/api/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newTestConn(t *testing.T) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return conn
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		store   grpcclient.TokenStore
		wantErr bool
	}{
		{
			name:    "nil store valid host",
			host:    "localhost:50051",
			store:   nil,
			wantErr: false,
		},
		{
			name:    "with store valid host",
			host:    "localhost:50051",
			store:   mocks.NewMockTokenStore(t),
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := grpcclient.NewClient(tc.host, tc.store)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, c)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, c)
			require.NoError(t, c.Close())
		})
	}
}

func TestClient_Register(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
	mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
	c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

	t.Run("success", func(t *testing.T) {
		mockAuth.EXPECT().
			Register(mock.Anything, &pb.RegisterRequest{Username: "user1", Password: "pass1"}).
			Return(&emptypb.Empty{}, nil)

		err := c.Register(context.Background(), "user1", "pass1")
		require.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		mockAuth.EXPECT().
			Register(mock.Anything, &pb.RegisterRequest{Username: "u", Password: "p"}).
			Return(nil, errors.New("username taken"))

		err := c.Register(context.Background(), "u", "p")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register")
	})
}

func TestClient_Login(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
	mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
	c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

	t.Run("success", func(t *testing.T) {
		mockAuth.EXPECT().
			Login(mock.Anything, &pb.LoginRequest{Username: "user2", Password: "pass2"}).
			Return(&pb.LoginResponse{Token: "jwt-token-123"}, nil)

		token, err := c.Login(context.Background(), "user2", "pass2")
		require.NoError(t, err)
		assert.Equal(t, "jwt-token-123", token)
	})

	t.Run("error from server", func(t *testing.T) {
		mockAuth.EXPECT().
			Login(mock.Anything, &pb.LoginRequest{Username: "u", Password: "p"}).
			Return(nil, errors.New("invalid credentials"))

		token, err := c.Login(context.Background(), "u", "p")
		require.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "failed to login")
	})
}

func TestClient_CreateItem(t *testing.T) {
	itemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "secret", Metadata: "meta"}}}

	t.Run("success", func(t *testing.T) {
		conn := newTestConn(t)
		defer conn.Close()
		mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
		mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
		c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

		mockItem.EXPECT().
			CreateItem(mock.Anything, &pb.CreateItemRequest{Type: pb.ItemType_TEXT, Data: itemData}).
			Return(&pb.CreateItemResponse{Item: &pb.Item{Id: "new-id", Type: pb.ItemType_TEXT}}, nil)

		err := c.CreateItem(context.Background(), pb.ItemType_TEXT, itemData)
		require.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		conn := newTestConn(t)
		defer conn.Close()
		mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
		mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
		c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

		mockItem.EXPECT().
			CreateItem(mock.Anything, mock.Anything).
			Return(nil, errors.New("create failed"))

		err := c.CreateItem(context.Background(), pb.ItemType_TEXT, itemData)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create item")
	})
}

func TestClient_UpdateItem(t *testing.T) {
	itemID := "item-uuid"
	itemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "updated", Metadata: "m"}}}
	updatedAt := timestamppb.Now()

	t.Run("success", func(t *testing.T) {
		conn := newTestConn(t)
		defer conn.Close()
		mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
		mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
		c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

		mockItem.EXPECT().
			UpdateItem(mock.Anything, &pb.UpdateItemRequest{Id: itemID, Type: pb.ItemType_TEXT, Data: itemData, UpdatedAt: updatedAt}).
			Return(&pb.UpdateItemResponse{Item: &pb.Item{Id: itemID}}, nil)

		err := c.UpdateItem(context.Background(), itemID, pb.ItemType_TEXT, itemData, updatedAt)
		require.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		conn := newTestConn(t)
		defer conn.Close()
		mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
		mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
		c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

		mockItem.EXPECT().
			UpdateItem(mock.Anything, mock.Anything).
			Return(nil, errors.New("conflict"))

		err := c.UpdateItem(context.Background(), itemID, pb.ItemType_TEXT, itemData, updatedAt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update item")
	})
}

func TestClient_DeleteItem(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
	mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
	c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

	t.Run("success", func(t *testing.T) {
		mockItem.EXPECT().
			DeleteItem(mock.Anything, &pb.DeleteItemRequest{Id: "item-id"}).
			Return(&emptypb.Empty{}, nil)

		err := c.DeleteItem(context.Background(), "item-id")
		require.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		mockItem.EXPECT().
			DeleteItem(mock.Anything, &pb.DeleteItemRequest{Id: "missing"}).
			Return(nil, errors.New("not found"))

		err := c.DeleteItem(context.Background(), "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete item")
	})
}

func TestClient_GetItem(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
	mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
	c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

	t.Run("success", func(t *testing.T) {
		want := &pb.Item{Id: "item-1", Type: pb.ItemType_CREDENTIAL}
		mockItem.EXPECT().
			GetItem(mock.Anything, &pb.GetItemRequest{Id: "item-1"}).
			Return(&pb.GetItemResponse{Item: want}, nil)

		got, err := c.GetItem(context.Background(), "item-1")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error from server", func(t *testing.T) {
		mockItem.EXPECT().
			GetItem(mock.Anything, &pb.GetItemRequest{Id: "missing"}).
			Return(nil, errors.New("not found"))

		got, err := c.GetItem(context.Background(), "missing")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to get item")
	})
}

func TestClient_ListItems(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	mockAuth := pbMocks.NewMockPasskeeperAuthServiceClient(t)
	mockItem := pbMocks.NewMockPasskeeperItemServiceClient(t)
	c := grpcclient.NewClientWithClients(conn, mockAuth, mockItem)

	t.Run("success", func(t *testing.T) {
		want := []*pb.Item{
			{Id: "id-1", Type: pb.ItemType_TEXT},
			{Id: "id-2", Type: pb.ItemType_CREDENTIAL},
		}
		mockItem.EXPECT().
			ListItems(mock.Anything, &pb.ListItemsRequest{Type: pb.ItemType_UNSPECIFIED}).
			Return(&pb.ListItemsResponse{Items: want}, nil)

		got, err := c.ListItems(context.Background(), pb.ItemType_UNSPECIFIED)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("success empty list", func(t *testing.T) {
		mockItem.EXPECT().
			ListItems(mock.Anything, &pb.ListItemsRequest{Type: pb.ItemType_CREDENTIAL}).
			Return(&pb.ListItemsResponse{Items: nil}, nil)

		got, err := c.ListItems(context.Background(), pb.ItemType_CREDENTIAL)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("error from server", func(t *testing.T) {
		mockItem.EXPECT().
			ListItems(mock.Anything, mock.Anything).
			Return(nil, errors.New("list failed"))

		got, err := c.ListItems(context.Background(), pb.ItemType_TEXT)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to list items")
	})
}

func TestClient_Close(t *testing.T) {
	conn := newTestConn(t)
	defer conn.Close()

	c := grpcclient.NewClientWithClients(conn, pbMocks.NewMockPasskeeperAuthServiceClient(t), pbMocks.NewMockPasskeeperItemServiceClient(t))
	err := c.Close()
	require.NoError(t, err)
}
