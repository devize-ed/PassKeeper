package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"passKeper/internal/auth"
	"passKeper/internal/repository/db"
	"passKeper/internal/server/mocks"
	"passKeper/internal/service"
	pb "passKeper/pkg/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAuthServer_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *pb.RegisterRequest
		setupMock   func(*mocks.MockAuthService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			req:  &pb.RegisterRequest{Username: "user1", Password: "pass123"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					CreateUser(ctx, "user1", "pass123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user already exists",
			req:  &pb.RegisterRequest{Username: "existing", Password: "pass"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					CreateUser(ctx, "existing", "pass").
					Return(db.ErrUserAlreadyExists)
			},
			wantErr:     true,
			wantCode:    codes.AlreadyExists,
			errContains: "already exists",
		},
		{
			name: "internal error",
			req:  &pb.RegisterRequest{Username: "user", Password: "pass"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					CreateUser(ctx, "user", "pass").
					Return(errors.New("db connection failed"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to register",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authSvc := mocks.NewMockAuthService(t)
			tc.setupMock(authSvc)
			srv := &AuthServer{AuthService: authSvc}

			resp, err := srv.Register(ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, &emptypb.Empty{}, resp)
		})
	}
}

func TestAuthServer_Login(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *pb.LoginRequest
		setupMock   func(*mocks.MockAuthService)
		wantErr     bool
		wantCode    codes.Code
		wantToken   string
		errContains string
	}{
		{
			name: "success",
			req:  &pb.LoginRequest{Username: "user1", Password: "pass123"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					LoginUser(ctx, "user1", "pass123").
					Return("jwt-token-xyz", nil)
			},
			wantErr:   false,
			wantToken: "jwt-token-xyz",
		},
		{
			name: "invalid credentials",
			req:  &pb.LoginRequest{Username: "user", Password: "wrong"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					LoginUser(ctx, "user", "wrong").
					Return("", service.ErrInvalidCredentials)
			},
			wantErr:     true,
			wantCode:    codes.Unauthenticated,
			errContains: "invalid credentials",
		},
		{
			name: "internal error",
			req:  &pb.LoginRequest{Username: "user", Password: "pass"},
			setupMock: func(m *mocks.MockAuthService) {
				m.EXPECT().
					LoginUser(ctx, "user", "pass").
					Return("", errors.New("service error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to login",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authSvc := mocks.NewMockAuthService(t)
			tc.setupMock(authSvc)
			srv := &AuthServer{AuthService: authSvc}

			resp, err := srv.Login(ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.wantToken, resp.Token)
		})
	}
}

func makeItemData() *pb.ItemData {
	return &pb.ItemData{
		Data: &pb.ItemData_Text{Text: &pb.Text{Text: "test data", Metadata: ""}},
	}
}

func makeDBItem(id, userID string, itemType int32, data *pb.ItemData) db.Item {
	now := time.Now()
	dataBytes, _ := proto.Marshal(data)
	return db.Item{
		ID:        id,
		Type:      itemType,
		Data:      dataBytes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestItemServer_CreateItem(t *testing.T) {
	userID := "testuser"
	ctx := authWithUserID(context.Background(), userID)
	itemData := makeItemData()
	itemType := int32(1)

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.CreateItemRequest
		setupMock   func(*mocks.MockItemService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.CreateItemRequest{Type: pb.ItemType_CREDENTIAL, Data: itemData},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					CreateItem(mock.Anything, itemType, mock.AnythingOfType("[]uint8")).
					Return(makeDBItem(uuid.New().String(), userID, itemType, itemData), nil)
			},
			wantErr: false,
		},
		{
			name: "invalid item type",
			ctx:  ctx,
			req:  &pb.CreateItemRequest{Type: pb.ItemType_UNSPECIFIED, Data: itemData},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					CreateItem(mock.Anything, int32(0), mock.AnythingOfType("[]uint8")).
					Return(db.Item{}, service.ErrInvalidItemType)
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "invalid item type",
		},
		{
			name: "internal error",
			ctx:  ctx,
			req:  &pb.CreateItemRequest{Type: pb.ItemType_CREDENTIAL, Data: itemData},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					CreateItem(mock.Anything, itemType, mock.AnythingOfType("[]uint8")).
					Return(db.Item{}, errors.New("storage error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to create item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemSvc := mocks.NewMockItemService(t)
			tc.setupMock(itemSvc)
			srv := &ItemServer{ItemService: itemSvc}

			resp, err := srv.CreateItem(tc.ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Item)
			assert.NotEmpty(t, resp.Item.Id)
			assert.Equal(t, tc.req.Type, resp.Item.Type)
		})
	}
}

func TestItemServer_UpdateItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := authWithUserID(context.Background(), userID)
	itemData := makeItemData()
	itemType := int32(2)
	ts := time.Now().UTC().Truncate(time.Microsecond)

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.UpdateItemRequest
		setupMock   func(*mocks.MockItemService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			ctx:  ctx,
			req: &pb.UpdateItemRequest{
				Id:        itemID,
				Type:      pb.ItemType_TEXT,
				Data:      itemData,
				UpdatedAt: timestamppb.New(ts),
			},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					UpdateItem(mock.Anything, itemID, itemType, mock.AnythingOfType("[]uint8"), ts).
					Return(makeDBItem(itemID, userID, itemType, itemData), nil)
			},
			wantErr: false,
		},
		{
			name: "item not found",
			ctx:  ctx,
			req: &pb.UpdateItemRequest{
				Id:        itemID,
				Type:      pb.ItemType_TEXT,
				Data:      itemData,
				UpdatedAt: timestamppb.New(ts),
			},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					UpdateItem(mock.Anything, itemID, itemType, mock.AnythingOfType("[]uint8"), ts).
					Return(db.Item{}, db.ErrItemNotFound)
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			errContains: "not found",
		},
		{
			name: "wrong user",
			ctx:  ctx,
			req: &pb.UpdateItemRequest{
				Id:        itemID,
				Type:      pb.ItemType_TEXT,
				Data:      itemData,
				UpdatedAt: timestamppb.New(ts),
			},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					UpdateItem(mock.Anything, itemID, itemType, mock.AnythingOfType("[]uint8"), ts).
					Return(db.Item{}, db.ErrWrongUserID)
			},
			wantErr:     true,
			wantCode:    codes.PermissionDenied,
			errContains: "does not belong",
		},
		{
			name: "timestamp too old",
			ctx:  ctx,
			req: &pb.UpdateItemRequest{
				Id:        itemID,
				Type:      pb.ItemType_TEXT,
				Data:      itemData,
				UpdatedAt: timestamppb.New(ts),
			},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					UpdateItem(mock.Anything, itemID, itemType, mock.AnythingOfType("[]uint8"), ts).
					Return(db.Item{}, db.ErrTimestampTooOld)
			},
			wantErr:     true,
			wantCode:    codes.FailedPrecondition,
			errContains: "timestamp",
		},
		{
			name: "internal error",
			ctx:  ctx,
			req: &pb.UpdateItemRequest{
				Id:        itemID,
				Type:      pb.ItemType_TEXT,
				Data:      itemData,
				UpdatedAt: timestamppb.New(ts),
			},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					UpdateItem(mock.Anything, itemID, itemType, mock.AnythingOfType("[]uint8"), ts).
					Return(db.Item{}, errors.New("storage error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to update",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemSvc := mocks.NewMockItemService(t)
			tc.setupMock(itemSvc)
			srv := &ItemServer{ItemService: itemSvc}

			resp, err := srv.UpdateItem(tc.ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Item)
			assert.Equal(t, itemID, resp.Item.Id)
		})
	}
}

func TestItemServer_DeleteItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := authWithUserID(context.Background(), userID)

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.DeleteItemRequest
		setupMock   func(*mocks.MockItemService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.DeleteItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					DeleteItem(ctx, itemID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "item not found",
			ctx:  ctx,
			req:  &pb.DeleteItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					DeleteItem(ctx, itemID).
					Return(db.ErrItemNotFound)
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			errContains: "not found",
		},
		{
			name: "wrong user",
			ctx:  ctx,
			req:  &pb.DeleteItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					DeleteItem(ctx, itemID).
					Return(db.ErrWrongUserID)
			},
			wantErr:     true,
			wantCode:    codes.PermissionDenied,
			errContains: "does not belong",
		},
		{
			name: "internal error",
			ctx:  ctx,
			req:  &pb.DeleteItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					DeleteItem(ctx, itemID).
					Return(errors.New("storage error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to delete",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemSvc := mocks.NewMockItemService(t)
			tc.setupMock(itemSvc)
			srv := &ItemServer{ItemService: itemSvc}

			resp, err := srv.DeleteItem(tc.ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, &emptypb.Empty{}, resp)
		})
	}
}

func TestItemServer_GetItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := authWithUserID(context.Background(), userID)
	itemData := makeItemData()
	dbItem := makeDBItem(itemID, userID, 1, itemData)

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.GetItemRequest
		setupMock   func(*mocks.MockItemService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.GetItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetItem(ctx, itemID).
					Return(dbItem, nil)
			},
			wantErr: false,
		},
		{
			name: "item not found",
			ctx:  ctx,
			req:  &pb.GetItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetItem(ctx, itemID).
					Return(db.Item{}, db.ErrItemNotFound)
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			errContains: "not found",
		},
		{
			name: "wrong user",
			ctx:  ctx,
			req:  &pb.GetItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetItem(ctx, itemID).
					Return(db.Item{}, db.ErrWrongUserID)
			},
			wantErr:     true,
			wantCode:    codes.PermissionDenied,
			errContains: "does not belong",
		},
		{
			name: "internal error",
			ctx:  ctx,
			req:  &pb.GetItemRequest{Id: itemID},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetItem(ctx, itemID).
					Return(db.Item{}, errors.New("storage error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to get item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemSvc := mocks.NewMockItemService(t)
			tc.setupMock(itemSvc)
			srv := &ItemServer{ItemService: itemSvc}

			resp, err := srv.GetItem(tc.ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Item)
			assert.Equal(t, itemID, resp.Item.Id)
			assert.NotNil(t, resp.Item.Data)
		})
	}
}

func TestItemServer_ListItems(t *testing.T) {
	userID := "testuser"
	ctx := authWithUserID(context.Background(), userID)
	itemData := makeItemData()
	items := []db.Item{
		makeDBItem("id1", userID, 1, itemData),
		makeDBItem("id2", userID, 2, itemData),
	}

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.ListItemsRequest
		setupMock   func(*mocks.MockItemService)
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "success",
			ctx:  ctx,
			req:  &pb.ListItemsRequest{Type: pb.ItemType_CREDENTIAL},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetAllItems(ctx, int32(1)).
					Return(items, nil)
			},
			wantErr: false,
		},
		{
			name: "item not found",
			ctx:  ctx,
			req:  &pb.ListItemsRequest{Type: pb.ItemType_CREDENTIAL},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetAllItems(ctx, int32(1)).
					Return(nil, db.ErrItemNotFound)
			},
			wantErr:     true,
			wantCode:    codes.NotFound,
			errContains: "not found",
		},
		{
			name: "wrong user",
			ctx:  ctx,
			req:  &pb.ListItemsRequest{Type: pb.ItemType_CREDENTIAL},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetAllItems(ctx, int32(1)).
					Return(nil, db.ErrWrongUserID)
			},
			wantErr:     true,
			wantCode:    codes.PermissionDenied,
			errContains: "does not belong",
		},
		{
			name: "internal error",
			ctx:  ctx,
			req:  &pb.ListItemsRequest{Type: pb.ItemType_CREDENTIAL},
			setupMock: func(m *mocks.MockItemService) {
				m.EXPECT().
					GetAllItems(ctx, int32(1)).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "failed to list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemSvc := mocks.NewMockItemService(t)
			tc.setupMock(itemSvc)
			srv := &ItemServer{ItemService: itemSvc}

			resp, err := srv.ListItems(tc.ctx, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
				if tc.errContains != "" {
					assert.Contains(t, st.Message(), tc.errContains)
				}
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Items, 2)
		})
	}
}

// authWithUserID puts userID in context imitating the auth interceptor.
func authWithUserID(ctx context.Context, userID string) context.Context {
	return auth.WithUserID(ctx, userID)
}
