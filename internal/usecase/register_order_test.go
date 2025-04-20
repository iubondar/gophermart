package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterOrderUsecase_RegisterOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test cases
	tests := []struct {
		name           string
		userID         uuid.UUID
		orderNumber    string
		repoResult     constants.OrderRegistrationResult
		repoError      error
		expectedResult constants.OrderRegistrationResult
		expectedError  error
	}{
		{
			name:           "Successful order registration",
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			repoResult:     constants.AcceptedToProcessing,
			expectedResult: constants.AcceptedToProcessing,
		},
		{
			name:           "Already registered order",
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			repoResult:     constants.AlreadyRegistered,
			expectedResult: constants.AlreadyRegistered,
		},
		{
			name:           "Wrong order number format",
			userID:         uuid.New(),
			orderNumber:    "invalid",
			repoResult:     constants.WrongOrderNumberFormat,
			expectedResult: constants.WrongOrderNumberFormat,
		},
		{
			name:           "Order registered by another user",
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			repoResult:     constants.RegisteredByAnotherUser,
			expectedResult: constants.RegisteredByAnotherUser,
		},
		{
			name:          "Repository error",
			userID:        uuid.New(),
			orderNumber:   "1234567890",
			repoError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := mocks.NewMockOrderRepository(ctrl)
			mockRepo.EXPECT().
				RegisterOrder(gomock.Any(), tt.userID, tt.orderNumber).
				Return(tt.repoResult, tt.repoError)

			// Create usecase
			uc := usecase.NewRegisterOrderUsecase(mockRepo)

			// Call usecase
			result, err := uc.RegisterOrder(context.Background(), tt.userID, tt.orderNumber)

			// Check results
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
