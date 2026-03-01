package hash_test

import (
	"strings"
	"testing"

	"passKeper/internal/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bcryptPrefix = "$2"

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 72),
			wantErr:  false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := hash.HashPassword(tc.password)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, got)
			assert.True(t, strings.HasPrefix(got, bcryptPrefix), "hash should be bcrypt format")

			got2, err := hash.HashPassword(tc.password)
			require.NoError(t, err)
			assert.NotEqual(t, got, got2, "bcrypt uses random salt")

			assert.True(t, hash.VerifyPassword(tc.password, got))
			assert.True(t, hash.VerifyPassword(tc.password, got2))
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	validHash, err := hash.HashPassword("secret")
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: "secret",
			hash:     validHash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrong",
			hash:     validHash,
			want:     false,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			hash:     validHash,
			want:     false,
		},
		{
			name:     "correct password empty hash",
			password: "secret",
			hash:     "",
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: "secret",
			hash:     "not-a-bcrypt-hash",
			want:     false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hash.VerifyPassword(tc.password, tc.hash)
			assert.Equal(t, tc.want, got)
		})
	}
}
