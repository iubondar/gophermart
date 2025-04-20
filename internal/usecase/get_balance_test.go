package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBalanceRepository struct {
	accountFn func(ctx context.Context, userID uuid.UUID) (models.Account, error)
}

func (m *mockBalanceRepository) Account(ctx context.Context, userID uuid.UUID) (models.Account, error) {
	return m.accountFn(ctx, userID)
}

func TestGetBalanceUsecase_GetBalance(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name          string
		repo          BalanceRepository
		expectedOut   BalanceOut
		expectedError error
	}{
		{
			name: "successful balance retrieval",
			repo: &mockBalanceRepository{
				accountFn: func(ctx context.Context, userID uuid.UUID) (models.Account, error) {
					return models.Account{
						Balance:       100.50,
						WithdrawalSum: 50.25,
					}, nil
				},
			},
			expectedOut: BalanceOut{
				Current:   100.50,
				Withdrawn: 50.25,
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			repo: &mockBalanceRepository{
				accountFn: func(ctx context.Context, userID uuid.UUID) (models.Account, error) {
					return models.Account{}, errors.New("database error")
				},
			},
			expectedOut:   BalanceOut{},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewGetBalanceUsecase(tt.repo)
			out, err := uc.GetBalance(ctx, userID)

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
