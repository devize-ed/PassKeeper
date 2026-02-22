package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"passKeper/internal/logger"
	"passKeper/internal/repository/migrations"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB struct represents a database connection.
type DB struct {
	pool *pgxpool.Pool
}

// Item struct represents an item instance in the database.
type Item struct {
	ID        string
	Type      int32
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewDB provides the new data base connection with the provided configuration.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	logger.Log.Debugf("Connecting to database with DSN: %s", dsn)

	// Run migrations before establishing the connection
	if err := migrations.RunMigrations(dsn, true); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	// Initialize a new connection pool with the provided DSN
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise a connection pool: %w", err)
	}

	logger.Log.Debug("Database connection established successfully")
	return &DB{
		pool: pool,
	}, nil
}

// initPool initializes a new connection pool.
func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	logger.Log.Debugf("Initializing a new connection pool with DSN: %s", dsn)
	// Parse the DSN and create a new connection pool with tracing enabled
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	// Set the connection pool configuration
	poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	// Ping the database to ensure the connection is established
	logger.Log.Debug("Pinging the database to ensure the connection is established")
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	logger.Log.Debug("Connection pool initialized successfully")
	return pool, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	logger.Log.Debug("Closing the database connection pool")
	db.pool.Close()
	logger.Log.Debug("Database connection pool closed successfully")
	return nil
}

// CreateUser creates a new user in the database.
func (db *DB) CreateUser(ctx context.Context, username string, passwordHash string) error {
	logger.Log.Debugf("Creating a new user with username: %s", username)
	// Generate a new UUID for the user
	id := uuid.New()
	// Insert the user into the database
	_, err := db.pool.Exec(ctx, `INSERT INTO users (id, username, password_hash, created_at)
	VALUES ($1, $2, $3, $4)`, id, username, passwordHash, time.Now().UTC())

	if err != nil {
		// If the username already exists, return an error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation { // if user already exists
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Log.Debugf("User created successfully with ID: %s", id)
	return nil
}

// GetUser gets a user from the database.
func (db *DB) GetUser(ctx context.Context, username string) (string, string, error) {
	logger.Log.Debugf("Getting a user with username: %s", username)
	// Query the user by username
	var userID, passwordHash string
	err := db.pool.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE username=$1", username).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", fmt.Errorf("failed to query the user: %w", err)
	}
	logger.Log.Debugf("User found successfully with ID: %s", userID)
	return userID, passwordHash, nil
}

// CreateItem creates a new item in the database.
func (db *DB) CreateItem(ctx context.Context, userID string, itemType int32, itemData []byte) (Item, error) {
	logger.Log.Debugf("Creating a new item with user ID: %s and item type: %d", userID, itemType)
	// Generate a new UUID for the item
	id := uuid.New()
	now := time.Now().UTC()
	// Insert the item into the database
	var item Item
	err := db.pool.QueryRow(ctx, `INSERT INTO items
	 (id, user_id, type, data, created_at, updated_at)
	 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, type, data, created_at, updated_at, deleted_at`,
		id, userID, itemType, itemData, now, now).Scan(&item.ID, &item.Type, &item.Data, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Item{}, ErrItemAlreadyExists
		}
		return Item{}, fmt.Errorf("failed to create the item: %w", err)
	}
	logger.Log.Debugf("Item created successfully with ID: %s", item.ID)
	return item, nil
}

// UpdateItem updates an item in the database and returns an updated item.
func (db *DB) UpdateItem(ctx context.Context, userID string, itemID string, itemType int32, itemData []byte, timestamp time.Time) (Item, error) {
	logger.Log.Debugf("Updating an item with user ID: %s and item ID: %s and item type: %d", userID, itemID, itemType)
	// Begin a transaction
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return Item{}, fmt.Errorf("failed to begin a transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	// Truncate the timestamp to the microsecond
	expectedTimestamp := timestamp.UTC().Truncate(time.Microsecond)
	// Get the new timestamp
	newTimestamp := time.Now().UTC().Truncate(time.Microsecond)

	// Update the item in the database and return the updated item
	var item Item
	err = tx.QueryRow(ctx,
		`UPDATE items SET data = $1, updated_at = $2 WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL AND type = $5 AND updated_at = $6 RETURNING id, type, data, created_at, updated_at, deleted_at`,
		itemData, newTimestamp, itemID, userID, itemType, expectedTimestamp,
	).Scan(&item.ID, &item.Type, &item.Data, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // if the item does not exist
			var exists int
			err = tx.QueryRow(ctx, "SELECT 1 FROM items WHERE id = $1 AND user_id = $2 AND type = $3 AND deleted_at IS NULL", itemID, userID, itemType).Scan(&exists)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return Item{}, ErrItemNotFound
				}
				return Item{}, fmt.Errorf("failed to check item existence: %w", err)
			}
			return Item{}, ErrTimestampTooOld
		}
		return Item{}, fmt.Errorf("failed to update the item: %w", err)
	}
	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return Item{}, fmt.Errorf("failed to commit the transaction: %w", err)
	}
	logger.Log.Debugf("Item updated successfully with ID: %s", item.ID)
	return item, nil
}

// GetItem gets an item from the database.
func (db *DB) GetItem(ctx context.Context, id string, userID string) (Item, error) {
	logger.Log.Debugf("Getting an item with ID: %s and user ID: %s", id, userID)
	// Query the item by id and user id
	var item Item
	err := db.pool.QueryRow(ctx, "SELECT id, type, data, created_at, updated_at, deleted_at FROM items WHERE id=$1 AND user_id=$2 AND deleted_at is null", id, userID).Scan(&item.ID, &item.Type, &item.Data, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Item{}, ErrItemNotFound
		}
		return Item{}, fmt.Errorf("failed to query the item: %w", err)
	}
	logger.Log.Debugf("Item found successfully with ID: %s", item.ID)
	return item, nil
}

// GetAllItems gets all items from the database for the given user ID, optionally filtered by itemType (0 = all types).
func (db *DB) GetAllItems(ctx context.Context, userID string, itemType int32) ([]Item, error) {
	logger.Log.Debugf("Getting all items for user ID: %s", userID)
	var qItems pgx.Rows
	var err error
	// If the item type is 0, get all items
	if itemType == 0 {
		qItems, err = db.pool.Query(ctx, "SELECT id, type, data, created_at, updated_at, deleted_at FROM items WHERE user_id=$1 AND deleted_at is null", userID)
		if err != nil {
			return nil, fmt.Errorf("failed to query the items: %w", err)
		}
		defer qItems.Close()
	} else {
		// If the item type is not 0, get all items of the given type
		qItems, err = db.pool.Query(ctx, "SELECT id, type, data, created_at, updated_at, deleted_at FROM items WHERE user_id=$1 AND type=$2 AND deleted_at is null", userID, itemType)
		if err != nil {
			return nil, fmt.Errorf("failed to query the items: %w", err)
		}
		defer qItems.Close()
	}
	// Scan the items data from the items
	var items []Item
	for qItems.Next() {
		var item Item
		err = qItems.Scan(&item.ID, &item.Type, &item.Data, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan the items: %w", err)
		}
		items = append(items, item)
	}
	if err := qItems.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan the items: %w", err)
	}
	logger.Log.Debugf("All items found successfully for user ID: %s", userID)
	return items, nil
}

// DeleteItem deletes an item from the database.
func (db *DB) DeleteItem(ctx context.Context, userID string, itemID string) error {
	logger.Log.Debugf("Deleting an item with user ID: %s and item ID: %s", userID, itemID)
	// Delete the item from the database
	now := time.Now().UTC()
	rowsAffected, err := db.pool.Exec(ctx, "UPDATE items SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND user_id = $4 AND deleted_at is null", now, now, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete the item: %w", err)
	}
	// If the item does not exist, return an error
	if rowsAffected.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	logger.Log.Debugf("Item deleted successfully with ID: %s", itemID)
	return nil
}
