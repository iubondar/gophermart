package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestPollingService_StartAndStop(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := mocks.NewMockOrderStatusFetcher(ctrl)
	repo := mocks.NewMockOrdersRepository(ctrl)

	// Setup expectations
	order := models.Order{
		UserID:     uuid.New(),
		Number:     "123",
		Status:     constants.OrderStatusNew,
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	orderStatus := models.OrderStatus{
		UserID:  order.UserID,
		Number:  order.Number,
		Status:  constants.OrderStatusProcessed,
		Accrual: 100.5,
	}

	// Expect OrdersToUpdate to be called multiple times
	repo.EXPECT().
		OrdersToUpdate(gomock.Any(), gomock.Any()).
		Return([]models.Order{order}, nil).
		MinTimes(1)

	// Expect FetchOrderStatus to be called for each order
	fetcher.EXPECT().
		FetchOrderStatus(gomock.Any(), order).
		Return(orderStatus, nil).
		MinTimes(1)

	// Expect UpdateOrders to be called with the processed status
	repo.EXPECT().
		UpdateOrders(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	// Create polling service with short interval for testing
	service := NewPollingService(100*time.Millisecond, 2, fetcher, repo)

	// Start the service
	service.Start()

	// Wait for a few polling cycles
	time.Sleep(300 * time.Millisecond)

	// Stop the service
	service.Stop()
}

func TestPollingService_ErrorHandling(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := mocks.NewMockOrderStatusFetcher(ctrl)
	repo := mocks.NewMockOrdersRepository(ctrl)

	// Setup test data
	order := models.Order{
		UserID:     uuid.New(),
		Number:     "123",
		Status:     constants.OrderStatusNew,
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	// Setup expectations
	repo.EXPECT().
		OrdersToUpdate(gomock.Any(), gomock.Any()).
		Return([]models.Order{order}, nil).
		MinTimes(1)

	fetcher.EXPECT().
		FetchOrderStatus(gomock.Any(), order).
		Return(models.OrderStatus{}, assert.AnError).
		MinTimes(1)

	repo.EXPECT().
		UpdateOrders(gomock.Any(), gomock.Any()).
		Return(assert.AnError).
		MinTimes(1)

	// Create polling service with short interval for testing
	service := NewPollingService(100*time.Millisecond, 2, fetcher, repo)

	// Start the service
	service.Start()

	// Wait for a few polling cycles
	time.Sleep(300 * time.Millisecond)

	// Stop the service
	service.Stop()
}

func TestPollingService_ContextCancellation(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := mocks.NewMockOrderStatusFetcher(ctrl)
	repo := mocks.NewMockOrdersRepository(ctrl)

	// Setup test data
	order := models.Order{
		UserID:     uuid.New(),
		Number:     "123",
		Status:     constants.OrderStatusNew,
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	orderStatus := models.OrderStatus{
		UserID:  order.UserID,
		Number:  order.Number,
		Status:  constants.OrderStatusProcessed,
		Accrual: 100.5,
	}

	// Setup expectations
	repo.EXPECT().
		OrdersToUpdate(gomock.Any(), gomock.Any()).
		Return([]models.Order{order}, nil).
		MinTimes(1)

	fetcher.EXPECT().
		FetchOrderStatus(gomock.Any(), order).
		DoAndReturn(func(ctx context.Context, order models.Order) (models.OrderStatus, error) {
			select {
			case <-ctx.Done():
				return models.OrderStatus{}, ctx.Err()
			default:
				return orderStatus, nil
			}
		}).
		MinTimes(1)

	repo.EXPECT().
		UpdateOrders(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	// Create polling service with short interval for testing
	service := NewPollingService(100*time.Millisecond, 2, fetcher, repo)

	// Start the service
	service.Start()

	// Wait for a few polling cycles to ensure at least one update
	time.Sleep(300 * time.Millisecond)

	// Stop the service
	service.Stop()
}

func TestPollingService_ConcurrentProcessing(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fetcher := mocks.NewMockOrderStatusFetcher(ctrl)
	repo := mocks.NewMockOrdersRepository(ctrl)

	// Setup test data
	orders := make([]models.Order, 5)
	for i := 0; i < 5; i++ {
		orders[i] = models.Order{
			UserID:     uuid.New(),
			Number:     string(rune('A' + i)),
			Status:     constants.OrderStatusNew,
			Accrual:    0,
			UploadedAt: time.Now(),
		}
	}

	// Setup expectations
	repo.EXPECT().
		OrdersToUpdate(gomock.Any(), gomock.Any()).
		Return(orders, nil).
		MinTimes(1)

	// Expect each order to be processed
	for _, order := range orders {
		orderStatus := models.OrderStatus{
			UserID:  order.UserID,
			Number:  order.Number,
			Status:  constants.OrderStatusProcessed,
			Accrual: 100.5,
		}
		fetcher.EXPECT().
			FetchOrderStatus(gomock.Any(), order).
			Return(orderStatus, nil).
			MinTimes(1)
	}

	repo.EXPECT().
		UpdateOrders(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	// Create polling service with multiple workers
	service := NewPollingService(100*time.Millisecond, 5, fetcher, repo)

	// Start the service
	service.Start()

	// Wait for a few polling cycles
	time.Sleep(300 * time.Millisecond)

	// Stop the service
	service.Stop()
}
