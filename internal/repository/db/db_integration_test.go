//go:build integration_tests
// +build integration_tests

package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code, err := runMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

const (
	testDBName       = "test"
	testUserName     = "test"
	testUserPassword = "test"
)

var getDSN func() string

func initGetDSN(hostAndPort string) {
	getDSN = func() string {
		return fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			testUserName,
			testUserPassword,
			hostAndPort,
			testDBName,
		)
	}
}

func getHostPort(hostPort string) (string, uint16, error) {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid host-port string: %s", hostPort)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port %s: %w", parts[1], err)
	}
	return parts[0], uint16(port), nil
}

func runMain(m *testing.M) (int, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 1, fmt.Errorf("failed to initialize dockertest pool: %w", err)
	}

	pg, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "17.2",
			Name:       "passkeeper-integration-tests",
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{"5432/tcp"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return 1, fmt.Errorf("failed to run postgres container: %w", err)
	}

	defer func() {
		if err := pool.Purge(pg); err != nil {
			log.Printf("failed to purge postgres container: %v", err)
		}
	}()

	hostPort := pg.GetHostPort("5432/tcp")
	initGetDSN(hostPort)

	pool.MaxWait = 10 * time.Second
	ctx := context.Background()

	var conn *pgx.Conn
	suDSN := fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", hostPort)
	if err := pool.Retry(func() error {
		var err error
		conn, err = pgx.Connect(ctx, suDSN)
		return err
	}); err != nil {
		return 1, fmt.Errorf("failed to connect to DB: %w", err)
	}

	defer func() {
		if conn != nil {
			_ = conn.Close(ctx)
		}
	}()

	if err := createTestDB(ctx, conn); err != nil {
		return 1, fmt.Errorf("failed to create test DB: %w", err)
	}

	return m.Run(), nil
}

func createTestDB(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, fmt.Sprintf(
		`CREATE USER %s WITH PASSWORD '%s'`,
		testUserName,
		testUserPassword,
	))
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	_, err = conn.Exec(ctx, fmt.Sprintf(
		`CREATE DATABASE %s OWNER %s ENCODING 'UTF8'`,
		testDBName, testUserName,
	))
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	return nil
}

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dsn := getDSN()
	d, err := NewDB(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	return d
}

func closeTestDB(t *testing.T, d *DB) {
	t.Helper()
	if err := d.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestDB_CreateUser(t *testing.T) {
	d := newTestDB(t)
	defer closeTestDB(t, d)

	cases := []struct {
		name      string
		username  string
		password  string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "create_user",
			username:  "user1",
			password:  "hash1",
			wantErr:   false,
			expectErr: nil,
		},
		{
			name:      "create_user_duplicate",
			username:  "user1",
			password:  "hash2",
			wantErr:   true,
			expectErr: ErrUserAlreadyExists,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := d.CreateUser(context.Background(), tc.username, tc.password)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDB_GetUser(t *testing.T) {
	d := newTestDB(t)
	defer closeTestDB(t, d)

	// Ensure user exists
	_ = d.CreateUser(context.Background(), "getuser1", "hash1")

	cases := []struct {
		name      string
		username  string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "get_user",
			username:  "getuser1",
			wantErr:   false,
			expectErr: nil,
		},
		{
			name:      "user_not_found",
			username:  "nonexistent",
			wantErr:   true,
			expectErr: ErrUserNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID, hash, err := d.GetUser(context.Background(), tc.username)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err)
				assert.Empty(t, userID)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, userID)
				assert.Equal(t, "hash1", hash)
			}
		})
	}
}

func TestDB_CreateItem_GetItem_GetAllItems(t *testing.T) {
	d := newTestDB(t)
	defer closeTestDB(t, d)

	_ = d.CreateUser(context.Background(), "itemuser1", "hash1")
	userID, _, err := d.GetUser(context.Background(), "itemuser1")
	require.NoError(t, err)
	require.NotEmpty(t, userID)

	itemType := int32(1)
	data := []byte("secret data")

	// CreateItem
	item, err := d.CreateItem(context.Background(), userID, itemType, data)
	require.NoError(t, err)
	require.NotEmpty(t, item.ID)
	assert.Equal(t, itemType, item.Type)
	assert.Equal(t, data, item.Data)

	// GetItem
	got, err := d.GetItem(context.Background(), item.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, itemType, got.Type)
	assert.Equal(t, data, got.Data)

	// GetAllItems (all types)
	items, err := d.GetAllItems(context.Background(), userID, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, item.ID, items[0].ID)

	// GetAllItems filtered by type
	items, err = d.GetAllItems(context.Background(), userID, itemType)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, item.ID, items[0].ID)

	// GetItem not found
	_, err = d.GetItem(context.Background(), "00000000-0000-0000-0000-000000000000", userID)
	assert.Error(t, err)
	assert.Equal(t, ErrItemNotFound, err)
}

func TestDB_UpdateItem(t *testing.T) {
	d := newTestDB(t)
	defer closeTestDB(t, d)

	_ = d.CreateUser(context.Background(), "updateuser1", "hash1")
	userID, _, _ := d.GetUser(context.Background(), "updateuser1")

	item, err := d.CreateItem(context.Background(), userID, 1, []byte("initial"))
	require.NoError(t, err)

	newData := []byte("updated data")
	updated, err := d.UpdateItem(context.Background(), userID, item.ID, 1, newData, item.UpdatedAt)
	require.NoError(t, err)
	assert.Equal(t, newData, updated.Data)

	// Wrong timestamp (optimistic lock)
	_, err = d.UpdateItem(context.Background(), userID, item.ID, 1, []byte("x"), item.UpdatedAt)
	assert.Error(t, err)
	assert.Equal(t, ErrTimestampTooOld, err)

	// Item not found
	_, err = d.UpdateItem(context.Background(), userID, "00000000-0000-0000-0000-000000000000", 1, []byte("x"), time.Now())
	assert.Error(t, err)
	assert.Equal(t, ErrItemNotFound, err)
}

func TestDB_DeleteItem(t *testing.T) {
	d := newTestDB(t)
	defer closeTestDB(t, d)

	_ = d.CreateUser(context.Background(), "deluser1", "hash1")
	userID, _, _ := d.GetUser(context.Background(), "deluser1")

	item, err := d.CreateItem(context.Background(), userID, 1, []byte("to delete"))
	require.NoError(t, err)

	err = d.DeleteItem(context.Background(), userID, item.ID)
	require.NoError(t, err)

	// GetItem returns not found after soft delete
	_, err = d.GetItem(context.Background(), item.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrItemNotFound, err)

	// DeleteItem again is not found
	err = d.DeleteItem(context.Background(), userID, item.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrItemNotFound, err)
}

func TestDB_Close(t *testing.T) {
	d := newTestDB(t)
	err := d.Close()
	assert.NoError(t, err)
}
