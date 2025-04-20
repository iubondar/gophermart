package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/iubondar/gophermart/internal/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrdersUsecase_GetOrders(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name           string
		mock           func(*mocks.MockOrdersRepository)
		expectedOrders []usecase.OrdersOut
		expectedErr    error
	}{
		{
			name: "successful retrieval with orders",
			mock: func(m *mocks.MockOrdersRepository) {
				m.EXPECT().Orders(ctx, userID).Return([]models.Order{
					{
						Number:     "12345678903",
						Status:     constants.OrderStatusProcessed,
						Accrual:    100.50,
						UploadedAt: now,
					},
					{
						Number:     "12345678904",
						Status:     constants.OrderStatusProcessing,
						Accrual:    0,
						UploadedAt: now,
					},
				}, nil)
			},
			expectedOrders: []usecase.OrdersOut{
				{
					Number:     "12345678903",
					Status:     constants.OrderStatusProcessed,
					Accrual:    100.50,
					UploadedAt: now.Format(time.RFC3339),
				},
				{
					Number:     "12345678904",
					Status:     constants.OrderStatusProcessing,
					Accrual:    0,
					UploadedAt: now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name: "successful retrieval with no orders",
			mock: func(m *mocks.MockOrdersRepository) {
				m.EXPECT().Orders(ctx, userID).Return([]models.Order{}, nil)
			},
			expectedOrders: []usecase.OrdersOut{},
			expectedErr:    nil,
		},
		{
			name: "repository error",
			mock: func(m *mocks.MockOrdersRepository) {
				m.EXPECT().Orders(ctx, userID).Return(nil, assert.AnError)
			},
			expectedOrders: nil,
			expectedErr:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrdersRepository(ctrl)
			tt.mock(mockRepo)

			uc := usecase.NewOrdersUsecase(mockRepo)
			orders, err := uc.GetOrders(ctx, userID)

			assert.Equal(t, tt.expectedOrders, orders)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
