package service

import (
	"context"
	"fmt"
	"passKeper/internal/auth"
	"passKeper/internal/logger"
	"passKeper/internal/repository/db"
	"time"
)

// AuthService struct provides the methods to interact with the auth service layer.
type ItemService struct {
	storage Storage
}

// NewItemService creates a new item service layer.
func NewItemService(storage Storage) *ItemService {
	return &ItemService{storage: storage}
}

// CreateItem creates a new item in the database.
func (s *ItemService) CreateItem(ctx context.Context, itemType int32, itemData []byte) (db.Item, error) {
	logger.Log.Debugf("Creating a new item with item type: %d", itemType)
	// Validate the item type
	err := validateItemType(itemType, false)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to validate item type: %w", err)
	}
	// Get the user ID from the token
	userID, err := auth.GetUserIDFromCtx(ctx)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to get user ID from token: %w", err)
	}
	// Create the item in the repository
	item, err := s.storage.CreateItem(ctx, userID, itemType, itemData)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to create item: %w", err)
	}
	return item, nil
}

// UpdateItem updates an item in the database.
func (s *ItemService) UpdateItem(ctx context.Context, itemID string, itemType int32, itemData []byte, timestamp time.Time) (db.Item, error) {
	logger.Log.Debugf("Updating an item with item ID: %s and item type: %d", itemID, itemType)
	// Validate the item type
	err := validateItemType(itemType, false)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to validate item type: %w", err)
	}
	// Get the user ID from the token
	userID, err := auth.GetUserIDFromCtx(ctx)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to get user ID from token: %w", err)
	}
	// Update the item in the repository
	item, err := s.storage.UpdateItem(ctx, userID, itemID, itemType, itemData, timestamp)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to update item: %w", err)
	}
	return item, nil
}

// DeleteItem deletes an item by ID.
func (s *ItemService) DeleteItem(ctx context.Context, itemID string) error {
	logger.Log.Debugf("Deleting an item with item ID: %s", itemID)
	// Get the user ID from the token
	userID, err := auth.GetUserIDFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user ID from token: %w", err)
	}
	// Delete the item in the repository
	err = s.storage.DeleteItem(ctx, userID, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// GetItem gets an item by ID.
func (s *ItemService) GetItem(ctx context.Context, id string) (db.Item, error) {
	logger.Log.Debugf("Getting an item with ID: %s", id)
	// Get the user ID from the token
	userID, err := auth.GetUserIDFromCtx(ctx)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to get user ID from token: %w", err)
	}
	// Get the item from the repository
	item, err := s.storage.GetItem(ctx, id, userID)
	if err != nil {
		return db.Item{}, fmt.Errorf("failed to get item: %w", err)
	}
	return item, nil
}

// GetAllItems gets all items.
func (s *ItemService) GetAllItems(ctx context.Context, itemType int32) ([]db.Item, error) {
	logger.Log.Debugf("Getting all items with item type: %d", itemType)
	// Get the user ID from the token
	userID, err := auth.GetUserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from token: %w", err)
	}
	if err := validateItemType(itemType, true); err != nil {
		return nil, fmt.Errorf("failed to validate item type: %w", err)
	}
	// Get the items from the repository
	items, err := s.storage.GetAllItems(ctx, userID, int32(itemType))
	if err != nil {
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}
	return items, nil
}

// validateItemType validates the item type.
func validateItemType(itemType int32, allowed bool) error {
	if itemType == 0 && allowed {
		return nil
	}
	if itemType < 1 || itemType > 4 {
		return fmt.Errorf("invalid item type: %d", itemType)
	}
	return nil
}
