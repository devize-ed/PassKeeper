package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"passKeper/internal/auth"
	"passKeper/internal/repository/db"
	"passKeper/internal/service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItemService(t *testing.T) {
	storage := mocks.NewMockStorage(t)
	svc := NewItemService(storage)
	require.NotNil(t, svc)
}

func testCtxWithUser(userID string) context.Context {
	return auth.WithUserID(context.Background(), userID)
}

func testMakeItem(id, userID string, itemType int32, data []byte) db.Item {
	now := time.Now()
	return db.Item{
		ID:        id,
		Type:      itemType,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestItemService_CreateItem(t *testing.T) {
	userID := "testuser"
	ctx := testCtxWithUser(userID)
	itemType := int32(1)
	itemData := []byte("secret data")

	tests := []struct {
		name        string
		ctx         context.Context
		itemType    int32
		itemData    []byte
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:     "success",
			ctx:      ctx,
			itemType: itemType,
			itemData: itemData,
			setupMock: func(m *mocks.MockStorage) {
				created := testMakeItem(uuid.New().String(), userID, itemType, itemData)
				m.EXPECT().
					CreateItem(ctx, userID, itemType, itemData).
					Return(created, nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid item type 0",
			ctx:         ctx,
			itemType:    0,
			itemData:    itemData,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "invalid item type 5",
			ctx:         ctx,
			itemType:    5,
			itemData:    itemData,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "missing user ID in context",
			ctx:         context.Background(),
			itemType:    itemType,
			itemData:    itemData,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "user ID",
		},
		{
			name:     "storage error",
			ctx:      ctx,
			itemType: itemType,
			itemData: itemData,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					CreateItem(ctx, userID, itemType, itemData).
					Return(db.Item{}, errors.New("db error"))
			},
			wantErr:     true,
			errContains: "create item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			svc := NewItemService(storage)

			item, err := svc.CreateItem(tc.ctx, tc.itemType, tc.itemData)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Equal(t, db.Item{}, item)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, tc.itemType, item.Type)
			assert.Equal(t, tc.itemData, item.Data)
		})
	}
}

func TestItemService_UpdateItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := testCtxWithUser(userID)
	itemType := int32(2)
	itemData := []byte("updated data")
	timestamp := time.Now()

	tests := []struct {
		name        string
		ctx         context.Context
		itemID      string
		itemType    int32
		itemData    []byte
		timestamp   time.Time
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:      "success",
			ctx:       ctx,
			itemID:    itemID,
			itemType:  itemType,
			itemData:  itemData,
			timestamp: timestamp,
			setupMock: func(m *mocks.MockStorage) {
				updated := testMakeItem(itemID, userID, itemType, itemData)
				m.EXPECT().
					UpdateItem(ctx, userID, itemID, itemType, itemData, timestamp).
					Return(updated, nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid item type",
			ctx:         ctx,
			itemID:      itemID,
			itemType:    0,
			itemData:    itemData,
			timestamp:   timestamp,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "missing user ID",
			ctx:         context.Background(),
			itemID:      itemID,
			itemType:    itemType,
			itemData:    itemData,
			timestamp:   timestamp,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "user ID",
		},
		{
			name:      "storage error",
			ctx:       ctx,
			itemID:    itemID,
			itemType:  itemType,
			itemData:  itemData,
			timestamp: timestamp,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					UpdateItem(ctx, userID, itemID, itemType, itemData, timestamp).
					Return(db.Item{}, errors.New("update failed"))
			},
			wantErr:     true,
			errContains: "update item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			svc := NewItemService(storage)

			item, err := svc.UpdateItem(tc.ctx, tc.itemID, tc.itemType, tc.itemData, tc.timestamp)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Equal(t, db.Item{}, item)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.itemID, item.ID)
			assert.Equal(t, tc.itemType, item.Type)
			assert.Equal(t, tc.itemData, item.Data)
		})
	}
}

func TestItemService_DeleteItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := testCtxWithUser(userID)

	tests := []struct {
		name        string
		ctx         context.Context
		itemID      string
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
			ctx:    ctx,
			itemID: itemID,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					DeleteItem(ctx, userID, itemID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "missing user ID",
			ctx:         context.Background(),
			itemID:      itemID,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "user ID",
		},
		{
			name:   "storage error",
			ctx:    ctx,
			itemID: itemID,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					DeleteItem(ctx, userID, itemID).
					Return(errors.New("delete failed"))
			},
			wantErr:     true,
			errContains: "delete",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			svc := NewItemService(storage)

			err := svc.DeleteItem(tc.ctx, tc.itemID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestItemService_GetItem(t *testing.T) {
	userID := "testuser"
	itemID := uuid.New().String()
	ctx := testCtxWithUser(userID)
	expectedItem := testMakeItem(itemID, userID, 1, []byte("data"))

	tests := []struct {
		name        string
		ctx         context.Context
		itemID      string
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
			ctx:    ctx,
			itemID: itemID,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetItem(ctx, itemID, userID).
					Return(expectedItem, nil)
			},
			wantErr: false,
		},
		{
			name:        "missing user ID",
			ctx:         context.Background(),
			itemID:      itemID,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "user ID",
		},
		{
			name:   "storage error",
			ctx:    ctx,
			itemID: itemID,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetItem(ctx, itemID, userID).
					Return(db.Item{}, errors.New("not found"))
			},
			wantErr:     true,
			errContains: "get item",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			svc := NewItemService(storage)

			item, err := svc.GetItem(tc.ctx, tc.itemID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Equal(t, db.Item{}, item)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, expectedItem.ID, item.ID)
			assert.Equal(t, expectedItem.Data, item.Data)
		})
	}
}

func TestItemService_GetAllItems(t *testing.T) {
	userID := "testuser"
	ctx := testCtxWithUser(userID)
	items := []db.Item{
		testMakeItem("id1", userID, 1, []byte("d1")),
		testMakeItem("id2", userID, 1, []byte("d2")),
	}

	tests := []struct {
		name        string
		ctx         context.Context
		itemType    int32
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:     "success with type 1",
			ctx:      ctx,
			itemType: 1,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetAllItems(ctx, userID, int32(1)).
					Return(items, nil)
			},
			wantErr: false,
		},
		{
			name:     "success with type 0 (all types)",
			ctx:      ctx,
			itemType: 0,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetAllItems(ctx, userID, int32(0)).
					Return(items, nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid item type 5",
			ctx:         ctx,
			itemType:    5,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "missing user ID",
			ctx:         context.Background(),
			itemType:    1,
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "user ID",
		},
		{
			name:     "storage error",
			ctx:      ctx,
			itemType: 1,
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetAllItems(ctx, userID, int32(1)).
					Return(nil, errors.New("db error"))
			},
			wantErr:     true,
			errContains: "get all items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			svc := NewItemService(storage)

			result, err := svc.GetAllItems(tc.ctx, tc.itemType)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tc.itemType != 0 {
				assert.Equal(t, items, result)
			} else {
				assert.Equal(t, items, result)
			}
		})
	}
}
