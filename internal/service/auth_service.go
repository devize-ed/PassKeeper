// Package service provides the business logic layer for PassKeeper.
// It implements authentication and item management on top of the storage layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"passKeper/internal/auth"
	"passKeper/internal/hash"
	"passKeper/internal/logger"
	"passKeper/internal/repository/db"
	"time"
)

// Sentinel errors returned by the service layer.
var (
	ErrInvalidCredentials = errors.New("invalid credentials") // ErrInvalidCredentials indicates wrong username or password.
	ErrInvalidItemType    = errors.New("invalid item type") // ErrInvalidItemType indicates an unsupported item type.
)

// Storage interface provides the methods to interact with the storage layer.
type Storage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) error
	GetUser(ctx context.Context, username string) (string, string, error)
	CreateItem(ctx context.Context, userID string, itemType int32, itemData []byte) (db.Item, error)
	UpdateItem(ctx context.Context, userID string, itemID string, itemType int32, itemData []byte, timestamp time.Time) (db.Item, error)
	DeleteItem(ctx context.Context, userID string, itemID string) error
	GetItem(ctx context.Context, id string, userID string) (db.Item, error)
	GetAllItems(ctx context.Context, userID string, itemType int32) ([]db.Item, error)
}

// Service struct provides the methods to interact with the service layer.
type AuthService struct {
	storage    Storage
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new auth service layer.
func NewAuthService(storage Storage, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{storage: storage, jwtManager: jwtManager}
}

// CreateUser creates a new user in the database.
func (s *AuthService) CreateUser(ctx context.Context, username string, password string) error {
	logger.Log.Debugf("Creating a new user with username: %s", username)
	// Validate the username and password
	err := auth.ValidateUser(username, password)
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	// Hash the password
	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	// Create the user in the repository
	err = s.storage.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// LoginUser logs in a user and returns a JWT token.
func (s *AuthService) LoginUser(ctx context.Context, username string, password string) (string, error) {
	logger.Log.Debugf("Logging in a user with username: %s", username)
	// Validate the username and password
	err := auth.ValidateUser(username, password)
	if err != nil {
		return "", fmt.Errorf("failed to validate user: %w", err)
	}
	// Get the user from the repository
	userID, passwordHash, err := s.storage.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	// Verify the password
	if !hash.VerifyPassword(password, passwordHash) {
		return "", fmt.Errorf("Invalid credentials: %w", ErrInvalidCredentials)
	}
	// Generate a new JWT token
	token, err := s.jwtManager.GenerateToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
