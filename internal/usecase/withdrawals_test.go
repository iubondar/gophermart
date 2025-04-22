package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/iubondar/gophermart/internal/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithdrawalsUsecase_GetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWithdrawalsRepository(ctrl)
	uc := usecase.NewWithdrawalsUsecase(mockRepo)

	userID := uuid.New()
	now := time.Now()

	testCases := []struct {
		name            string
		repoWithdrawals []models.Withdrawal
		repoError       error
		expected        []usecase.WithdrawalOut
		expectedError   error
	}{
		{
			name: "successful withdrawal retrieval",
			repoWithdrawals: []models.Withdrawal{
				{
					Number:      "1234567890",
					Sum:         100.50,
					ProcessedAt: now,
				},
			},
			expected: []usecase.WithdrawalOut{
				{
					Order:       "1234567890",
					Sum:         100.50,
					ProcessedAt: now.Format(time.RFC3339),
				},
			},
		},
		{
			name:            "no withdrawals",
			repoWithdrawals: []models.Withdrawal{},
			expected:        []usecase.WithdrawalOut{},
		},
		{
			name:          "repository error",
			repoError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.EXPECT().
				Withdrawals(gomock.Any(), userID).
				Return(tc.repoWithdrawals, tc.repoError)

			result, err := uc.GetWithdrawals(context.Background(), userID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
