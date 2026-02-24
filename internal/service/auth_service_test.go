package service

import (
	"context"
	"errors"
	"testing"

	"passKeper/internal/auth"
	"passKeper/internal/hash"
	"passKeper/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key-for-jwt"

func TestNewAuthService(t *testing.T) {
	storage := mocks.NewMockStorage(t)
	jwtManager := auth.NewJWTManager(testJWTSecret)
	svc := NewAuthService(storage, jwtManager)
	require.NotNil(t, svc)
}

func TestAuthService_CreateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		username    string
		password    string
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
	}{
		{
			name:     "success",
			username: "testuser",
			password: "validpass123",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					CreateUser(ctx, "testuser", mock.AnythingOfType("string")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "empty username",
			username:    "",
			password:    "pass",
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "empty password",
			username:    "user",
			password:    "",
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:     "storage create user error",
			username: "testuser",
			password: "validpass123",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					CreateUser(ctx, "testuser", mock.AnythingOfType("string")).
					Return(errors.New("user already exists"))
			},
			wantErr:     true,
			errContains: "create user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			jwtManager := auth.NewJWTManager(testJWTSecret)
			svc := NewAuthService(storage, jwtManager)

			err := svc.CreateUser(ctx, tc.username, tc.password)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	ctx := context.Background()
	correctHash, err := testHashPassword("mypassword123")
	require.NoError(t, err)
	wrongHash, err := testHashPassword("otherpassword")
	require.NoError(t, err)

	tests := []struct {
		name        string
		username    string
		password    string
		setupMock   func(*mocks.MockStorage)
		wantErr     bool
		errContains string
		wantErrIs   error
		wantUserID  string
	}{
		{
			name:     "success",
			username: "testuser",
			password: "mypassword123",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetUser(ctx, "testuser").
					Return("user-id-123", correctHash, nil)
			},
			wantUserID: "user-id-123",
		},
		{
			name:     "invalid credentials",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetUser(ctx, "testuser").
					Return("user-id-123", wrongHash, nil)
			},
			wantErr:     true,
			wantErrIs:   ErrInvalidCredentials,
			errContains: "credentials",
		},
		{
			name:        "empty username",
			username:    "",
			password:    "pass",
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:        "empty password",
			username:    "user",
			password:    "",
			setupMock:   func(m *mocks.MockStorage) {},
			wantErr:     true,
			errContains: "validate",
		},
		{
			name:     "get user error",
			username: "testuser",
			password: "validpass123",
			setupMock: func(m *mocks.MockStorage) {
				m.EXPECT().
					GetUser(ctx, "testuser").
					Return("", "", errors.New("user not found"))
			},
			wantErr:     true,
			errContains: "get user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)
			tc.setupMock(storage)
			jwtManager := auth.NewJWTManager(testJWTSecret)
			svc := NewAuthService(storage, jwtManager)

			token, err := svc.LoginUser(ctx, tc.username, tc.password)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			if tc.wantUserID != "" {
				userID, parseErr := jwtManager.ParseToken(token)
				require.NoError(t, parseErr)
				assert.Equal(t, tc.wantUserID, userID)
			}
		})
	}
}

func testHashPassword(pwd string) (string, error) {
	return hash.HashPassword(pwd)
}
