package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLoginRepository struct {
	checkLoginFn func(ctx context.Context, login string, password string) (userID uuid.UUID, err error)
}

func (m *mockLoginRepository) CheckLogin(ctx context.Context, login string, password string) (userID uuid.UUID, err error) {
	return m.checkLoginFn(ctx, login, password)
}

func TestLoginUsecase_Login(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name          string
		repo          LoginRepository
		login         string
		password      string
		expectedOut   uuid.UUID
		expectedError error
	}{
		{
			name: "successful login",
			repo: &mockLoginRepository{
				checkLoginFn: func(ctx context.Context, login string, password string) (uuid.UUID, error) {
					return userID, nil
				},
			},
			login:         "testuser",
			password:      "testpass",
			expectedOut:   userID,
			expectedError: nil,
		},
		{
			name: "repository error",
			repo: &mockLoginRepository{
				checkLoginFn: func(ctx context.Context, login string, password string) (uuid.UUID, error) {
					return uuid.Nil, errors.New("database error")
				},
			},
			login:         "testuser",
			password:      "testpass",
			expectedOut:   uuid.Nil,
			expectedError: errors.New("database error"),
		},
		{
			name: "invalid credentials",
			repo: &mockLoginRepository{
				checkLoginFn: func(ctx context.Context, login string, password string) (uuid.UUID, error) {
					return uuid.Nil, nil
				},
			},
			login:         "testuser",
			password:      "wrongpass",
			expectedOut:   uuid.Nil,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewLoginUsecase(tt.repo)
			out, err := uc.Login(ctx, tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOut, out)
			}
		})
	}
}
