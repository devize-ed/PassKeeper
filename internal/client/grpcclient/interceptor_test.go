package grpcclient

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockTokenStore implements TokenStore for tests.
type mockTokenStore struct {
	token string
	err   error
}

func (m *mockTokenStore) Load() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestAuthInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		store         TokenStore
		method        string
		wantInvoked   bool
		wantAuthValue string // expected "Bearer <token>" in outgoing context, or "" if none
		wantErr       bool
		errContains   string
	}{
		{
			name:          "nil store",
			store:         nil,
			method:        "/some/Service/Method",
			wantInvoked:   true,
			wantAuthValue: "",
			wantErr:       false,
		},
		{
			name:          "login method",
			store:         &mockTokenStore{err: errors.New("no token")},
			method:        LoginMethod,
			wantInvoked:   true,
			wantAuthValue: "",
			wantErr:       false,
		},
		{
			name:          "register method",
			store:         &mockTokenStore{err: errors.New("no token")},
			method:        RegisterMethod,
			wantInvoked:   true,
			wantAuthValue: "",
			wantErr:       false,
		},
		{
			name:          "auth method",
			store:         &mockTokenStore{token: "jwt-token"},
			method:        "/passkeeper.PasskeeperItemService/GetItem",
			wantInvoked:   true,
			wantAuthValue: "Bearer jwt-token",
			wantErr:       false,
		},
		{
			name:          "empty token",
			store:         &mockTokenStore{token: ""},
			method:        "/passkeeper.PasskeeperItemService/ListItems",
			wantInvoked:   true,
			wantAuthValue: "",
			wantErr:       false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var invoked bool
			var capturedMD metadata.MD
			fakeInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				invoked = true
				md, ok := metadata.FromOutgoingContext(ctx)
				if ok {
					capturedMD = md.Copy()
				}
				return nil
			}

			interceptor := authInterceptor(tc.store)
			ctx := context.Background()
			err := interceptor(ctx, tc.method, nil, nil, nil, fakeInvoker)

			assert.Equal(t, tc.wantInvoked, invoked)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			require.NoError(t, err)
			if tc.wantAuthValue != "" {
				require.NotNil(t, capturedMD)
				auth := capturedMD.Get("authorization")
				require.Len(t, auth, 1)
				assert.Equal(t, tc.wantAuthValue, auth[0])
			}
		})
	}
}
