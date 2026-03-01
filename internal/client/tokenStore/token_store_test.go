package tokenStore_test

import (
	"os"
	"path/filepath"
	"testing"

	"passKeper/internal/client/tokenStore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	ts := tokenStore.NewTokenStore(path)
	require.NotNil(t, ts)
	assert.Equal(t, path, ts.TokenPath)

	tsHome := tokenStore.NewTokenStore("~/.passkeeper/token")
	require.NotNil(t, tsHome)
	if tsHome.TokenPath != "~/.passkeeper/token" {
		assert.Contains(t, tsHome.TokenPath, ".passkeeper")
	}
}

func TestTokenStore_Load(t *testing.T) {
	tests := []struct {
		name        string
		fileContent *string
		wantToken   string
		wantErr     bool
		errContains string
	}{

		{
			name:        "valid token",
			fileContent: strPtr("my-jwt-token"),
			wantToken:   "my-jwt-token",
			wantErr:     false,
		},
		{
			name:        "file not exist",
			fileContent: nil,
			wantErr:     true,
			errContains: "login",
		},
		{
			name:        "empty file",
			fileContent: strPtr(""),
			wantErr:     true,
			errContains: "login",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token")
			if tc.fileContent != nil {
				err := os.WriteFile(path, []byte(*tc.fileContent), 0600)
				require.NoError(t, err)
			}
			ts := tokenStore.NewTokenStore(path)

			loaded, err := ts.Load()
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, loaded)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, loaded)
		})
	}
}

func TestTokenStore_Save(t *testing.T) {
	tests := []struct {
		name       string
		pathSuffix string
		tokens     []string
		wantToken  string
		wantErr    bool
	}{
		{
			name:       "roundtrip creates dir and file",
			pathSuffix: filepath.Join("subdir", "token"),
			tokens:     []string{"my-jwt-token-here"},
			wantToken:  "my-jwt-token-here",
			wantErr:    false,
		},
		{
			name:       "creates nested directory",
			pathSuffix: filepath.Join("a", "b", "c", "token"),
			tokens:     []string{"token-value"},
			wantToken:  "token-value",
			wantErr:    false,
		},
		{
			name:       "overwrites existing",
			pathSuffix: "token",
			tokens:     []string{"first", "second"},
			wantToken:  "second",
			wantErr:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.pathSuffix)
			ts := tokenStore.NewTokenStore(path)

			for _, token := range tc.tokens {
				err := ts.Save(token)
				if tc.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			if tc.wantErr {
				return
			}
			_, err := os.Stat(path)
			require.NoError(t, err)
			loaded, err := ts.Load()
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, loaded)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
