package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test cases
	tests := []struct {
		name           string
		in             usecase.RegisterIn
		repoOk         bool
		repoError      error
		expectedUserID uuid.UUID
		expectedOk     bool
		expectedError  error
	}{
		{
			name: "Successful registration",
			in: usecase.RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			repoOk:         true,
			expectedOk:     true,
			expectedUserID: uuid.New(),
		},
		{
			name: "Empty login",
			in: usecase.RegisterIn{
				Login:    "",
				Password: "testpass",
			},
			expectedOk: false,
		},
		{
			name: "Empty password",
			in: usecase.RegisterIn{
				Login:    "testuser",
				Password: "",
			},
			expectedOk: false,
		},
		{
			name: "User already exists",
			in: usecase.RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			repoOk:     false,
			expectedOk: false,
		},
		{
			name: "Repository error",
			in: usecase.RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			repoError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := mocks.NewMockUserRepository(ctrl)
			if tt.in.Login != "" && tt.in.Password != "" {
				mockRepo.EXPECT().
					Register(gomock.Any(), gomock.Any(), tt.in.Login, gomock.Any()).
					Return(tt.repoOk, tt.repoError)
			}

			// Create usecase
			uc := usecase.NewRegisterUsecase(mockRepo)

			// Call usecase
			userID, ok, err := uc.Register(context.Background(), tt.in)

			// Check results
			if tt.expectedUserID != uuid.Nil {
				assert.NotEqual(t, uuid.Nil, userID)
			}
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
