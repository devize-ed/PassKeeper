package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"passKeper/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func TestJWTManager_GenerateToken(t *testing.T) {
	j := auth.NewJWTManager(testSecret)

	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"valid user id", "testuser", false},
		{"empty user id", "", false},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := j.GenerateToken(tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, token)
			// JWT format: header.payload.signature
			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3)
			//parse and check user_id
			parsed, err := j.ParseToken(token)
			require.NoError(t, err)
			assert.Equal(t, tc.userID, parsed)
		})
	}
}

func TestJWTManager_ParseToken(t *testing.T) {
	j := auth.NewJWTManager(testSecret)
	validToken, err := j.GenerateToken("parsed-user")
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user",
		"iat":     jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		"exp":     jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	jWrong := auth.NewJWTManager("wrong-secret")

	tests := []struct {
		name    string
		manager *auth.JWTManager
		token   string
		wantID  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			manager: j,
			token:   validToken,
			wantID:  "parsed-user",
			wantErr: false,
		},
		{
			name:    "wrong secret",
			manager: jWrong,
			token:   validToken,
			wantErr: true,
			errMsg:  "signature",
		},
		{
			name:    "expired token",
			manager: j,
			token:   expiredToken,
			wantErr: true,
			errMsg:  "expired",
		},
		{
			name:    "malformed token",
			manager: j,
			token:   "not.a.jwt",
			wantErr: true,
		},
		{
			name:    "empty token",
			manager: j,
			token:   "",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := tc.manager.ParseToken(tc.token)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, userID)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantID, userID)
		})
	}
}

func TestWithUserID_GetUserIDFromCtx(t *testing.T) {
	ctx := context.Background()

	// Context without user ID
	_, err := auth.GetUserIDFromCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID")

	// set then get
	userID := "ctx-user-1"
	ctxWith := auth.WithUserID(ctx, userID)
	got, err := auth.GetUserIDFromCtx(ctxWith)
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	// override the user ID
	ctxWith2 := auth.WithUserID(ctxWith, "override")
	got2, err := auth.GetUserIDFromCtx(ctxWith2)
	require.NoError(t, err)
	assert.Equal(t, "override", got2)
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid",
			username: "user",
			password: "pass",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			password: "pass",
			wantErr:  true,
			errMsg:   "required",
		},
		{
			name:     "empty password",
			username: "user",
			password: "",
			wantErr:  true,
			errMsg:   "required",
		},
		{
			name:     "both empty",
			username: "",
			password: "",
			wantErr:  true,
			errMsg:   "required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := auth.ValidateUser(tc.username, tc.password)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}
