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

func TestWithdrawUsecase_Withdraw(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name           string
		in             usecase.WithdrawIn
		mock           func(*mocks.MockWithdrawer)
		expectedResult constants.WithdrawResult
		expectedErr    error
	}{
		{
			name: "successful withdrawal",
			in: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			mock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(ctx, userID, "12345678903", float32(100.50)).
					Return(constants.Success, nil)
			},
			expectedResult: constants.Success,
			expectedErr:    nil,
		},
		{
			name: "insufficient funds",
			in: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			mock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(ctx, userID, "12345678903", float32(100.50)).
					Return(constants.InsufficientFunds, nil)
			},
			expectedResult: constants.InsufficientFunds,
			expectedErr:    nil,
		},
		{
			name: "empty order number",
			in: usecase.WithdrawIn{
				Order: "",
				Sum:   100.50,
			},
			mock:           func(m *mocks.MockWithdrawer) {},
			expectedResult: constants.WrongOrderFormat,
			expectedErr:    nil,
		},
		{
			name: "zero sum",
			in: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   0,
			},
			mock:           func(m *mocks.MockWithdrawer) {},
			expectedResult: constants.WrongOrderFormat,
			expectedErr:    nil,
		},
		{
			name: "negative sum",
			in: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   -100.50,
			},
			mock:           func(m *mocks.MockWithdrawer) {},
			expectedResult: constants.WrongOrderFormat,
			expectedErr:    nil,
		},
		{
			name: "withdrawer error",
			in: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			mock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(ctx, userID, "12345678903", float32(100.50)).
					Return(constants.Success, assert.AnError)
			},
			expectedResult: constants.Success,
			expectedErr:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWithdrawer := mocks.NewMockWithdrawer(ctrl)
			tt.mock(mockWithdrawer)

			uc := usecase.NewWithdrawUsecase(mockWithdrawer)
			result, err := uc.Withdraw(ctx, userID, tt.in)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
