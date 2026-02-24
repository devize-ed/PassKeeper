// Package tokenStore provides file-based JWT token persistence for the client.
package tokenStore

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Errors
var (
	errNotLoggedIn     = errors.New("Authentication token is empty, please login")
	errFileInteraction = errors.New("file interaction failed")
)

// TokenStore struct provides the methods to interact with the token store.
type TokenStore struct {
	TokenPath string
}

// NewTokenStore creates a new token store instance.
func NewTokenStore(tokenStorePath string) *TokenStore {
	// expand the token store path if it contains ~
	path := tokenStorePath
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				path = homeDir
			} else if path == "~/" || strings.HasPrefix(path, "~/") {
				path = filepath.Join(homeDir, path[2:])
			}
		}
	}
	return &TokenStore{TokenPath: path}
}

// Load reads the token from the token store file.
func (t *TokenStore) Load() (string, error) {
	// check if the token store directory exists
	if _, err := os.Stat(t.TokenPath); os.IsNotExist(err) {
		return "", errNotLoggedIn
	}
	// read the token from the token store file
	tokenBytes, err := os.ReadFile(t.TokenPath)
	if err != nil {
		return "", errFileInteraction
	}
	// trim spaces/newlines
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", errNotLoggedIn
	}
	return token, nil
}

// Save writes the token to the token store file.
func (t *TokenStore) Save(token string) error {
	// create the token store directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(t.TokenPath), 0700)
	if err != nil {
		return errFileInteraction
	}
	// write the token to the token store file
	err = os.WriteFile(t.TokenPath, []byte(token), 0600)
	if err != nil {
		return errFileInteraction
	}
	return nil
}
