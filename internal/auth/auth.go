// Package auth provides JWT-based authentication utilities for the PassKeeper service.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager is the manager for the JWT authentication.
type JWTManager struct {
	secret []byte // secret for the JWT authentication.
}

// ctxKey is the type of the context key.
type ctxKey int

// userIDKey is the key for the user ID in the context.
const userIDKey ctxKey = iota

var (
	errClaimNotFound      = errors.New("user_id not found in claims")     // errClaimNotFound is the error returned when the user ID is not found in the claims.
	errCredRequired       = errors.New("login and password are required") // errCredRequired is the error returned when the login and password are required.
	errInvalidTokenMethod = errors.New("invalid token method")            // errInvalidTokenMethod is the error returned when the token method is invalid.
	errUserIDNotFound     = errors.New("user ID not found in context")    // errUserIDNotFound is the error returned when the user ID is not found in the context.
	errInvalidToken       = errors.New("invalid token")
	errTokenExpired       = errors.New("token expired") // errTokenExpired is the error returned when the token is expired.
)

// NewJWTManager creates a JWT manager with the given signing secret.
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

// GenerateToken generates a new JWT token for the user.
func (j *JWTManager) GenerateToken(userID string) (string, error) {
	// Generate a new JWT token for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iat":     jwt.NewNumericDate(time.Now()),
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	// Sign the token with the secret
	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// ParseToken parses a JWT token and returns the user ID.
func (j *JWTManager) ParseToken(tokenStr string) (userID string, err error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// verify the token method is HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("invalid token method: %w", errInvalidTokenMethod)
		}
		return j.secret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token: %w", errInvalidToken)
	}
	// get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to get claims from token: %w", err)
	}
	// get the expiration time from the claims
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", fmt.Errorf("expiration time is not a float64")
	}
	// check if the expiration time is in the past
	if time.Now().After(time.Unix(int64(exp), 0)) {
		return "", errTokenExpired
	}
	// if the user ID is not found in the claims, return an error
	if claims["user_id"] == nil {
		return "", errClaimNotFound
	}
	userID, ok = claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user ID is not a string: %w", errInvalidToken)
	}
	return userID, nil
}

// WithUserID adds the user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromCtx gets the user ID from the context.
func GetUserIDFromCtx(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", errUserIDNotFound
	}
	return userID, nil
}

// ValidateUser validates the username and password are non-empty.
func ValidateUser(username string, password string) error {
	// If the username or password is empty, return an error
	if username == "" || password == "" {
		return errCredRequired
	}
	return nil
}
