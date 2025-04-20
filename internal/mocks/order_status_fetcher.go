package mocks

import (
	"context"

	"github.com/iubondar/gophermart/internal/models"
)

//go:generate mockgen -source=order_status_fetcher.go -destination=./mock_order_status_fetcher.go -package=mocks
type OrderStatusFetcher interface {
	FetchOrderStatus(ctx context.Context, order models.Order) (out models.OrderStatus, err error)
}
