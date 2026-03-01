package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	client "passKeper/internal/client/grpcclient"
	"passKeper/internal/hash"
	"passKeper/internal/repository/db"
	serverpkg "passKeper/internal/server"
	"passKeper/internal/service/mocks"
	pb "passKeper/pkg/api"

	"passKeper/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockTokenStore struct {
	token string
	err   error
}

func (m *mockTokenStore) Save(token string) error {
	m.token = token
	return nil
}
func (m *mockTokenStore) Load() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func mustNewClient(t *testing.T, addr string) *client.Client {
	t.Helper()
	c, err := client.NewClient(addr, nil, false, "", "")
	require.NoError(t, err)
	return c
}

func mustNewClientWithToken(t *testing.T, addr string, token string) *client.Client {
	t.Helper()
	ts := &mockTokenStore{token: token}
	c, err := client.NewClient(addr, ts, false, "", "")
	require.NoError(t, err)
	return c
}

func TestNewApp(t *testing.T) {
	c := mustNewClient(t, "localhost:50051")
	defer func() { _ = c.Close() }()

	app, err := NewApp(c, &mockTokenStore{token: "t"})
	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestApp_Close(t *testing.T) {
	c := mustNewClient(t, "localhost:50051")
	app, err := NewApp(c, nil)
	require.NoError(t, err)
	require.NoError(t, app.Close())
}

func TestApp_requireAuthentication(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		loadErr     error
		wantErr     bool
		errContains string
	}{
		{"has token", "valid-token", nil, false, ""},
		{"empty token", "", nil, true, "authentication required"},
		{"store load error", "", errors.New("load failed"), true, "authentication required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{
				grpcClient: nil,
				ts:         &mockTokenStore{token: tc.token, err: tc.loadErr},
				in:         bufio.NewReader(bytes.NewReader(nil)),
				out:        bufio.NewWriter(&bytes.Buffer{}),
			}
			err := a.requireAuthentication()
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApp_Register(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().CreateUser(mock.Anything, "user1", mock.AnythingOfType("string")).Return(nil)

	jwt := auth.NewJWTManager("test-secret")
	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClient(t, addr)
	defer func() { _ = c.Close() }()
	app, err := NewApp(c, &mockTokenStore{})
	require.NoError(t, err)

	err = app.Register(ctx, "user1", "pass1")
	require.NoError(t, err)
}

func TestApp_Login(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	password := "pass123"
	passwordHash, err := hash.HashPassword(password)
	require.NoError(t, err)

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().GetUser(mock.Anything, "user1").Return("user-id", passwordHash, nil)

	jwt := auth.NewJWTManager("test-secret")
	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClient(t, addr)
	defer func() { _ = c.Close() }()
	ts := &mockTokenStore{}
	app, err := NewApp(c, ts)
	require.NoError(t, err)

	err = app.Login(ctx, "user1", password)
	require.NoError(t, err)
	assert.NotEmpty(t, ts.token)
}

func TestApp_CreateItem(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().CreateItem(mock.Anything, "user-id", int32(pb.ItemType_TEXT), mock.AnythingOfType("[]uint8")).Return(db.Item{ID: "item-1", Type: int32(pb.ItemType_TEXT)}, nil)

	jwt := auth.NewJWTManager("test-secret")
	token, err := jwt.GenerateToken("user-id")
	require.NoError(t, err)

	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClientWithToken(t, addr, token)
	defer func() { _ = c.Close() }()

	// stdin: text + metadata for type 2 (text)
	in := bufio.NewReader(bytes.NewReader([]byte("mytext\nmymeta\n")))
	app := &App{
		grpcClient: c,
		ts:         &mockTokenStore{token: token},
		in:         in,
		out:        bufio.NewWriter(&bytes.Buffer{}),
	}

	err = app.CreateItem(ctx, 2)
	require.NoError(t, err)
}

func TestApp_GetItem(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	itemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "x", Metadata: "y"}}}
	dataBytes, err := proto.Marshal(itemData)
	require.NoError(t, err)

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().GetItem(mock.Anything, "item-1", "user-id").Return(db.Item{
		ID:        "item-1",
		Type:      int32(pb.ItemType_TEXT),
		Data:      dataBytes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	jwt := auth.NewJWTManager("test-secret")
	token, err := jwt.GenerateToken("user-id")
	require.NoError(t, err)

	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClientWithToken(t, addr, token)
	defer func() { _ = c.Close() }()

	app, err := NewApp(c, &mockTokenStore{token: token})
	require.NoError(t, err)
	app.in = bufio.NewReader(bytes.NewReader(nil))
	app.out = bufio.NewWriter(&bytes.Buffer{})

	item, err := app.GetItem(ctx, "item-1")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "item-1", item.Id)
	assert.Equal(t, pb.ItemType_TEXT, item.Type)
}

func TestApp_ListItems(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	itemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "a", Metadata: "b"}}}
	dataBytes, err := proto.Marshal(itemData)
	require.NoError(t, err)

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().GetAllItems(mock.Anything, "user-id", int32(0)).Return([]db.Item{{
		ID:        "item-1",
		Type:      int32(pb.ItemType_TEXT),
		Data:      dataBytes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}, nil)

	jwt := auth.NewJWTManager("test-secret")
	token, err := jwt.GenerateToken("user-id")
	require.NoError(t, err)

	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClientWithToken(t, addr, token)
	defer func() { _ = c.Close() }()

	app, err := NewApp(c, &mockTokenStore{token: token})
	require.NoError(t, err)
	app.in = bufio.NewReader(bytes.NewReader(nil))
	app.out = bufio.NewWriter(&bytes.Buffer{})

	err = app.ListItems(ctx, 0)
	require.NoError(t, err)
}

func TestApp_EditItem(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	_ = lis.Close()

	itemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "old", Metadata: "oldmeta"}}}
	dataBytes, err := proto.Marshal(itemData)
	require.NoError(t, err)
	newItemData := &pb.ItemData{Data: &pb.ItemData_Text{Text: &pb.Text{Text: "new", Metadata: "newmeta"}}}
	newDataBytes, err := proto.Marshal(newItemData)
	require.NoError(t, err)
	updatedAt := time.Now()

	storage := mocks.NewMockStorage(t)
	storage.EXPECT().GetItem(mock.Anything, "item-1", "user-id").Return(db.Item{
		ID:        "item-1",
		Type:      int32(pb.ItemType_TEXT),
		Data:      dataBytes,
		CreatedAt: time.Now(),
		UpdatedAt: updatedAt,
	}, nil)
	storage.EXPECT().UpdateItem(mock.Anything, "user-id", "item-1", int32(pb.ItemType_TEXT), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Time")).Return(db.Item{
		ID:        "item-1",
		Type:      int32(pb.ItemType_TEXT),
		Data:      newDataBytes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	jwt := auth.NewJWTManager("test-secret")
	token, err := jwt.GenerateToken("user-id")
	require.NoError(t, err)

	srv := serverpkg.NewServer(jwt, storage, false, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx, addr, jwt) }()
	time.Sleep(50 * time.Millisecond)

	c := mustNewClientWithToken(t, addr, token)
	defer func() { _ = c.Close() }()

	// stdin: new text + metadata for formItemData(type 2)
	in := bufio.NewReader(bytes.NewReader([]byte("new\nnewmeta\n")))
	app := &App{
		grpcClient: c,
		ts:         &mockTokenStore{token: token},
		in:         in,
		out:        bufio.NewWriter(&bytes.Buffer{}),
	}

	err = app.EditItem(ctx, "item-1")
	require.NoError(t, err)
}
